package rvasp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/envelope"
	"github.com/trisacrypto/trisa/pkg/trisa/mtls"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm"
)

// TRISA implements the GRPC TRISANetwork and TRISAHealth services.
type TRISA struct {
	protocol.UnimplementedTRISANetworkServer
	protocol.UnimplementedTRISAHealthServer
	parent *Server
	srv    *grpc.Server
	certs  *trust.Provider
	chain  trust.ProviderPool
	sign   *rsa.PrivateKey
}

// NewTRISA from a parent server.
func NewTRISA(parent *Server) (svc *TRISA, err error) {
	conf := parent.conf
	if conf.CertPath == "" || conf.TrustChainPath == "" {
		return nil, errors.New("certificate path or trust chain path missing")
	}

	var sz *trust.Serializer
	svc = &TRISA{parent: parent}

	// Load the TRISA certificates for server-side TLS
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, err
	}

	if svc.certs, err = sz.ReadFile(conf.CertPath); err != nil {
		return nil, err
	}

	// Load the TRISA public pool from disk
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, err
	}

	if svc.chain, err = sz.ReadPoolFile(conf.TrustChainPath); err != nil {
		return nil, err
	}

	// Extract the signing key from the TRISA certificate
	// TODO: use separate signing key from mTLS certs
	if svc.sign, err = svc.certs.GetRSAKeys(); err != nil {
		return nil, err
	}
	return svc, nil
}

// Serve initializes the GRPC server and returns any errors during intitialization, it
// then kicks off a go routine to handle requests. Not thread safe, should not be called
// multiple times.
func (s *TRISA) Serve() (err error) {
	// Create TLS Credentials for the server
	// NOTE: the mtls package specifies TRISA-specific TLS configuration.
	var creds grpc.ServerOption
	if creds, err = mtls.ServerCreds(s.certs, s.chain); err != nil {
		return err
	}

	// Create a new gRPC server with panic recovery and tracing middleware
	s.srv = grpc.NewServer(creds, grpc.UnaryInterceptor(UnaryTraceInterceptor), grpc.StreamInterceptor(StreamTraceInterceptor))
	protocol.RegisterTRISANetworkServer(s.srv, s)
	protocol.RegisterTRISAHealthServer(s.srv, s)

	var sock net.Listener
	if sock, err = net.Listen("tcp", s.parent.conf.TRISABindAddr); err != nil {
		return fmt.Errorf("trisa service could not listen on %q", s.parent.conf.TRISABindAddr)
	}

	go s.AsyncHandler(nil)

	go s.Run(sock)

	log.Info().
		Str("listen", s.parent.conf.TRISABindAddr).
		Msg("trisa server started")

	return nil
}

// Run the gRPC server. This method is extracted from the Serve function so that it can
// be run in its own go routine and to allow tests to Run a bufconn server without
// starting a live server with all of the various go routines and channels running.
func (s *TRISA) Run(sock net.Listener) {
	defer sock.Close()
	if err := s.srv.Serve(sock); err != nil {
		s.parent.echan <- err
	}
}

// Shutdown the TRISA server gracefully
func (s *TRISA) Shutdown() (err error) {
	log.Info().Msg("trisa server gracefully shutting down")
	s.srv.GracefulStop()
	log.Debug().Msg("successful trisa server shutdown")
	return nil
}

// Transfer enables a quick one-off transaction between peers.
func (s *TRISA) Transfer(ctx context.Context, in *protocol.SecureEnvelope) (out *protocol.SecureEnvelope, err error) {
	// Get the peer from the context
	var peer *peers.Peer
	if peer, err = s.parent.peers.FromContext(ctx); err != nil {
		log.Error().Err(err).Msg("could not verify peer from context")
		return nil, &protocol.Error{
			Code:    protocol.Unverified,
			Message: err.Error(),
		}
	}
	log.Info().Str("peer", peer.String()).Msg("unary transfer request received")
	s.parent.updates.Broadcast(0, fmt.Sprintf("received secure exchange from %s", peer), pb.MessageCategory_TRISAP2P)

	// Check signing key is available to send an encrypted response
	if peer.SigningKey() == nil {
		log.Warn().Str("peer", peer.String()).Msg("no remote signing key available, attempting key exchange")
		s.parent.updates.Broadcast(0, "no remote signing key available, attempting key exchange", pb.MessageCategory_TRISAP2P)

		if _, err = peer.ExchangeKeys(false); err != nil {
			log.Warn().Err(err).Str("peer", peer.String()).Msg("no remote signing key available, key exchange failed")
			s.parent.updates.Broadcast(0, fmt.Sprintf("key exchange failed: %s", err), pb.MessageCategory_TRISAP2P)
		}

		// Second check for signing keys, if they're not available then reject messages
		if peer.SigningKey() == nil {
			return nil, &protocol.Error{
				Code:    protocol.NoSigningKey,
				Message: "please retry transfer after key exchange",
				Retry:   true,
			}
		}
	}

	return s.handleTransaction(ctx, peer, in)
}

// TransferStream allows for high-throughput transactions.
func (s *TRISA) TransferStream(stream protocol.TRISANetwork_TransferStreamServer) (err error) {
	// Get the peer from the context
	ctx := stream.Context()
	var peer *peers.Peer
	if peer, err = s.parent.peers.FromContext(ctx); err != nil {
		log.Error().Err(err).Msg("could not verify peer from context")
		return &protocol.Error{
			Code:    protocol.Unverified,
			Message: err.Error(),
		}
	}
	log.Info().Str("peer", peer.String()).Msg("transfer stream opened")
	s.parent.updates.Broadcast(0, fmt.Sprintf("transfer stream opened from %s", peer), pb.MessageCategory_TRISAP2P)

	// Check signing key is available to send an encrypted response
	if peer.SigningKey() == nil {
		log.Warn().Str("peer", peer.String()).Msg("no remote signing key available, attempting key exchange")
		s.parent.updates.Broadcast(0, "no remote signing key available, attempting key exchange", pb.MessageCategory_TRISAP2P)

		if _, err = peer.ExchangeKeys(false); err != nil {
			log.Warn().Err(err).Str("peer", peer.String()).Msg("no remote signing key available, key exchange failed")
			s.parent.updates.Broadcast(0, fmt.Sprintf("key exchange failed: %s", err), pb.MessageCategory_TRISAP2P)
		}

		// Second check for signing keys, if they're not available then reject messages
		if peer.SigningKey() == nil {
			return &protocol.Error{
				Code:    protocol.NoSigningKey,
				Message: "please retry transfer after key exchange",
				Retry:   true,
			}
		}
	}

	// Handle incoming secure envelopes from client
	// TODO: add go routines to parallelize handling rather than one transfer at a time
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var in *protocol.SecureEnvelope
		if in, err = stream.Recv(); err == io.EOF {
			log.Info().Str("peer", peer.String()).Msg("transfer stream closed")
		} else if err != nil {
			log.Warn().Err(err).Msg("recv stream error")
			return protocol.Errorf(protocol.Unavailable, "stream closed prematurely: %s", err)
		}

		// Handle the response
		out, err := s.handleTransaction(ctx, peer, in)
		if err != nil {
			// Do not close the stream, send the error in the secure envelope if the
			// Error is a TRISA coded error.
			pbErr, ok := err.(*protocol.Error)
			if !ok {
				return err
			}
			out = &protocol.SecureEnvelope{
				Error: pbErr,
			}
		}

		if err = stream.Send(out); err != nil {
			log.Error().Err(err).Msg("send stream error")
			return err
		}
		log.Info().Str("peer", peer.String()).Str("id", in.Id).Msg("streaming transfer request complete")
	}
}

func (s *TRISA) handleTransaction(ctx context.Context, peer *peers.Peer, in *protocol.SecureEnvelope) (out *protocol.SecureEnvelope, err error) {
	var (
		identity    *ivms101.IdentityPayload
		transaction *generic.Transaction
	)

	// Check for TRISA rejection errors
	reject, isErr := envelope.Check(in)
	if isErr {
		if reject != nil {
			if out, err = envelope.Reject(reject, envelope.WithEnvelopeID(in.Id)); err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
			}
			return out, nil
		}
		log.Error().Err(err).Msg("TRISA protocol error while checking envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Check the algorithms to make sure they're supported
	if in.EncryptionAlgorithm != "AES256-GCM" || in.HmacAlgorithm != "HMAC-SHA256" {
		log.Warn().
			Str("encryption", in.EncryptionAlgorithm).
			Str("hmac", in.HmacAlgorithm).
			Msg("unsupported cryptographic algorithms")
		s.parent.updates.Broadcast(0, "server only supports AES256-GCM and HMAC-SHA256", pb.MessageCategory_TRISAP2P)
		return nil, protocol.Errorf(protocol.UnhandledAlgorithm, "server only supports AES256-GCM and HMAC-SHA256")
	}
	s.parent.updates.Broadcast(0, "decrypting with RSA and AES256-GCM; verifying with HMAC-SHA256", pb.MessageCategory_TRISAP2P)

	// Decrypt the encryption key and HMAC secret with private signing keys (asymmetric phase)
	payload, reject, err := envelope.Open(in, envelope.WithRSAPrivateKey(s.sign))
	if err != nil {
		if reject != nil {
			if out, err = envelope.Reject(reject, envelope.WithEnvelopeID(in.Id)); err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
			}
			return out, nil
		}
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	if identity, transaction, _, err = parsePayload(payload, false); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while parsing payload")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	s.parent.updates.Broadcast(0, fmt.Sprintf("secure envelope %s opened and payload decrypted and parsed", in.Id), pb.MessageCategory_TRISAP2P)

	// Check if we are the originator of the transaction
	var localIdentity *ivms101.Person
	if localIdentity, err = s.parent.vasp.LoadIdentity(); err != nil {
		log.Warn().Err(err).Msg("could not load local VASP identity")
		return nil, protocol.Errorf(protocol.InternalError, "could not load local VASP identity")
	}

	// For async transactions the originator receives a transfer request from the
	// beneficiary, so call the originator handler to continue the transaction.
	if proto.Equal(localIdentity, identity.OriginatingVasp.OriginatingVasp) {
		// We need the remote peer's endpoint in order to initiate async transactions
		// with the beneficiary.
		if peer, err = s.parent.peers.Search(peer.String()); err != nil {
			log.Warn().Err(err).Msg("could not lookup remote peer")
			return nil, protocol.Errorf(protocol.InternalError, "could not lookup remote peer")
		}
		return s.parent.respondAsync(peer, payload, identity, transaction, in.Id)
	}

	// Lookup the beneficiary in the local VASP database.
	var accountAddress string
	if transaction.Beneficiary == "" {
		log.Warn().Msg("no beneficiary information supplied")
		return nil, protocol.Errorf(protocol.MissingFields, "beneficiary wallet address or email required in transaction")
	} else {
		accountAddress = transaction.Beneficiary
	}

	var account db.Account
	if err = s.parent.db.LookupAccount(accountAddress).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Str("account", accountAddress).Msg("unknown beneficiary")
			return nil, protocol.Errorf(protocol.UnkownBeneficiary, "could not find beneficiary %q", accountAddress)
		}
		log.Error().Err(err).Msg("could not lookup beneficiary account")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Retrieve the wallet for the beneficiary account
	var wallet db.Wallet
	if err = s.parent.db.LookupWallet(account.WalletAddress).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Str("wallet", account.WalletAddress).Msg("unknown beneficiary wallet")
			return nil, protocol.Errorf(protocol.UnkownWalletAddress, "could not find beneficiary wallet %q", account.WalletAddress)
		}
		log.Error().Err(err).Msg("could not lookup beneficiary wallet")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Run the scenario for the wallet's configured policy
	policy := wallet.BeneficiaryPolicy
	log.Debug().Str("wallet", account.WalletAddress).Str("policy", string(policy)).Msg("received transfer request")
	switch policy {
	case db.SyncRepair:
		// Respond to the transfer request immediately, filling in the beneficiary
		// identity information.
		return s.respondTransfer(in, peer, identity, transaction, account, false)
	case db.SyncRequire:
		// Respond to the transfer request immediately, requiring that the beneficiary
		// identity is already filled in.
		return s.respondTransfer(in, peer, identity, transaction, account, true)
	case db.AsyncRepair:
		// Respond to the transfer request with a pending message and mark the
		// transaction for later service. The beneficiary information is filled in.
		return s.respondPending(in, peer, identity, transaction, account)
	case db.AsyncReject:
		// Respond to the transfer request with a pending message that will be later
		// rejected.
		return s.respondPending(in, peer, identity, transaction, account)
	default:
		return nil, protocol.Errorf(protocol.InternalError, "unknown policy '%s' for wallet '%s'", policy, account.WalletAddress)
	}
}

// Repair the identity payload in a received transfer request by filling in the
// beneficiary identity information.
func (s *TRISA) repairBeneficiary(identity *ivms101.IdentityPayload, account db.Account) (err error) {
	identity.BeneficiaryVasp = &ivms101.BeneficiaryVasp{}
	if identity.BeneficiaryVasp.BeneficiaryVasp, err = s.parent.vasp.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load beneficiary vasp")
		return err
	}

	identity.Beneficiary = &ivms101.Beneficiary{
		BeneficiaryPersons: make([]*ivms101.Person, 0, 1),
		AccountNumbers:     []string{account.WalletAddress},
	}

	var beneficiary *ivms101.Person
	if beneficiary, err = account.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load beneficiary account identity")
		return err
	}
	identity.Beneficiary.BeneficiaryPersons = append(identity.Beneficiary.BeneficiaryPersons, beneficiary)
	return nil
}

// respondTransfer responds to a transfer request from the originator by sending back
// the payload with the beneficiary identity information. If requireBeneficiary is
// true, the beneficiary identity must be filled in, or the transfer is rejected. If
// requireBeneficiary is false, the partial beneficiary identity is repaired.
func (s *TRISA) respondTransfer(in *protocol.SecureEnvelope, peer *peers.Peer, identity *ivms101.IdentityPayload, transaction *generic.Transaction, account db.Account, requireBeneficiary bool) (out *protocol.SecureEnvelope, err error) {
	// Save the pending transaction on the account
	account.Pending++
	if err = s.parent.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save beneficiary account")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	if requireBeneficiary {
		// TODO: Validate the actual fields in the beneficiary identity
		if identity.BeneficiaryVasp == nil || identity.BeneficiaryVasp.BeneficiaryVasp == nil {
			log.Error().Msg("TRISA protocol error: missing beneficiary vasp identity")
			return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary vasp identity")
		}
	} else {
		// Fill in the beneficiary identity information for the repair policy
		s.repairBeneficiary(identity, account)
	}

	// Update the transaction with beneficiary information
	transaction.Beneficiary = account.WalletAddress
	transaction.Timestamp = time.Now().Format(time.RFC3339)

	// Save the completed transaction in the database
	var ach db.Transaction
	if ach, err = s.parent.db.MakeTransaction(transaction.Originator, transaction.Beneficiary); err != nil {
		log.Error().Err(err).Msg("could not create transaction")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}
	ach.Envelope = in.Id
	ach.Account = account
	ach.Amount = decimal.NewFromFloat(transaction.Amount)
	ach.Debit = false
	ach.State = pb.TransactionState_COMPLETED

	var achBytes []byte
	if achBytes, err = protojson.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal transaction identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}
	ach.Identity = string(achBytes)

	if err = s.parent.db.Create(&ach).Error; err != nil {
		log.Error().Err(err).Msg("could not create transaction in database")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Update the account information
	account.Balance = account.Balance.Add(decimal.NewFromFloat(transaction.Amount))
	account.Completed++
	account.Pending--
	if err = s.parent.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save beneficiary account")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	msg := fmt.Sprintf("ready for transaction %04d: %s transfering from %s to %s", ach.ID, ach.Amount, ach.Originator.WalletAddress, ach.Beneficiary.WalletAddress)
	s.parent.updates.Broadcast(0, msg, pb.MessageCategory_BLOCKCHAIN)

	// Encode and encrypt the payload information to return the secure envelope
	payload := &protocol.Payload{
		SentAt:     time.Now().Format(time.RFC3339),
		ReceivedAt: time.Now().Format(time.RFC3339),
	}
	if payload.Identity, err = anypb.New(identity); err != nil {
		log.Error().Err(err).Msg("could not dump payload identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}
	if payload.Transaction, err = anypb.New(transaction); err != nil {
		log.Error().Err(err).Msg("could not dump payload transaction")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	s.parent.updates.Broadcast(0, "sealing beneficiary information and returning", pb.MessageCategory_TRISAP2P)

	out, reject, err := envelope.Seal(payload, envelope.WithRSAPublicKey(peer.SigningKey()))
	if err != nil {
		if reject != nil {
			if out, err = envelope.Reject(reject, envelope.WithEnvelopeID(in.Id)); err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
			}
			return out, nil
		}
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	s.parent.updates.Broadcast(0, fmt.Sprintf("%04d new account balance: %s", account.ID, account.Balance), pb.MessageCategory_LEDGER)
	return out, nil
}

// respondPending responds to a transfer request from the originator by returning a
// pending message and saving the pending transaction in the database.
func (s *TRISA) respondPending(in *protocol.SecureEnvelope, peer *peers.Peer, identity *ivms101.IdentityPayload, transaction *generic.Transaction, account db.Account) (out *protocol.SecureEnvelope, err error) {
	now := time.Now()

	// Check if the pending transaction exists, otherwise create it
	var xfer db.Transaction
	if err = s.parent.db.LookupTransaction(in.Id).First(&xfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new pending transaction in the database
			if xfer, err = s.parent.db.MakeTransaction(transaction.Originator, transaction.Beneficiary); err != nil {
				log.Error().Err(err).Msg("could not construct transaction")
				return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
			}
			xfer.Envelope = in.Id
			xfer.Account = account
			xfer.Amount = decimal.NewFromFloat(transaction.Amount)
			xfer.Debit = false
			xfer.State = pb.TransactionState_PENDING_SENT

			if err = s.parent.db.Create(&xfer).Error; err != nil {
				log.Error().Err(err).Msg("could not create transaction in database")
				return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
			}
		} else {
			log.Error().Err(err).Msg("could not perform transaction lookup")
			return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
		}
	} else {
		// Update the existing pending transaction
		xfer.State = pb.TransactionState_PENDING_ACKNOWLEDGED
	}

	xfer.NotBefore = now.Add(s.parent.conf.AsyncNotBefore)
	xfer.NotAfter = now.Add(s.parent.conf.AsyncNotAfter)

	// Marshal the identity info into the local transaction
	var data []byte
	if data, err = protojson.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}
	xfer.Identity = string(data)

	// Marshal the generic.Transaction into the local transaction
	if data, err = protojson.Marshal(transaction); err != nil {
		log.Error().Err(err).Msg("could not marshal transaction")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}
	xfer.Transaction = string(data)

	// Save the updated transaction in the database
	if err = s.parent.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Save the pending account in the database
	account.Pending++
	if err = s.parent.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not update beneficiary account")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Cache the common name of the originator in the database for later retrieval
	var originator *db.Identity
	if originator, err = xfer.GetOriginator(s.parent.db); err != nil {
		log.Error().Err(err).Msg("could not get originator identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	originator.Provider = peer.Info().CommonName
	if err = s.parent.db.Save(&originator).Error; err != nil {
		log.Error().Err(err).Msg("could not update originator identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Create a pending protocol message with NotBefore and NotAfter timestamps
	pending := &generic.Pending{
		EnvelopeId:     xfer.Envelope,
		ReplyNotBefore: xfer.NotBefore.Format(time.RFC3339),
		ReplyNotAfter:  xfer.NotAfter.Format(time.RFC3339),
	}

	var payload *protocol.Payload
	if payload, err = createPendingPayload(pending, identity); err != nil {
		log.Error().Err(err).Msg("could not create pending payload")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	out, reject, err := envelope.Seal(payload, envelope.WithRSAPublicKey(peer.SigningKey()))
	if err != nil {
		if reject != nil {
			if out, err = envelope.Reject(reject, envelope.WithEnvelopeID(in.Id)); err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
			}
			return out, nil
		}
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	return out, nil
}

// sendAsync handles a pending transaction in the database by performing an
// envelope transfer with the originator and updating the database accordingly.
func (s *TRISA) sendAsync(tx *db.Transaction) (err error) {
	// Fetch the originator address
	var originator *db.Identity
	if originator, err = tx.GetOriginator(s.parent.db); err != nil {
		log.Error().Err(err).Msg("could not fetch originator address")
		return fmt.Errorf("could not fetch originator address")
	}

	// Create the identity for the payload
	identity := &ivms101.IdentityPayload{}
	if err = protojson.Unmarshal([]byte(tx.Identity), identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity from transaction")
		return fmt.Errorf("could not unmarshal identity from transaction: %s", err)
	}

	// Repair the beneficiary information if this is the first handshake
	if tx.State == pb.TransactionState_PENDING_SENT {
		var account *db.Account
		if account, err = tx.GetAccount(s.parent.db); err != nil {
			log.Error().Err(err).Msg("could not fetch beneficiary account")
			return fmt.Errorf("could not fetch beneficiary account: %s", err)
		}

		if err = s.repairBeneficiary(identity, *account); err != nil {
			log.Error().Err(err).Msg("could not repair beneficiary information")
			return fmt.Errorf("could not repair beneficiary information: %s", err)
		}
	}

	// Create the generic.Transaction for the payload
	transaction := &generic.Transaction{}
	if err = protojson.Unmarshal([]byte(tx.Transaction), transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal generic.Transaction from transaction")
		return fmt.Errorf("could not unmarshal generic.Transaction from transaction: %s", err)
	}

	// Create the payload
	var payload *protocol.Payload
	if payload, err = createTransferPayload(identity, transaction); err != nil {
		log.Error().Err(err).Msg("could not create transfer payload")
		return fmt.Errorf("could not create transfer payload: %s", err)
	}
	payload.ReceivedAt = time.Now().Format(time.RFC3339)

	// Conduct a GDS lookup if necessary to get the endpoint
	var peer *peers.Peer
	if peer, err = s.parent.peers.Search(originator.Provider); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return fmt.Errorf("could not search peer from directory service: %s", err)
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return fmt.Errorf("could not exchange keys with remote peer: %s", err)
	}

	// Secure the envelope with the remote originator's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(tx.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transfer, handle errors
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return fmt.Errorf("could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.sign))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	if _, transaction, _, err = parsePayload(payload, true); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while parsing payload")
		return fmt.Errorf("TRISA protocol error while parsing payload: %s", err)
	}

	if transaction == nil {
		// We expected an echo from the counterparty to conclude an async but got back
		// a pending or other type of correctly parsed response.
		log.Warn().
			Str("transaction_type", payload.Transaction.TypeUrl).
			Msg("unexpected transaction reply to async completion")
		return fmt.Errorf("received %q payload expected a generic Transaction echo", payload.Transaction.TypeUrl)
	}

	switch tx.State {
	case pb.TransactionState_PENDING_SENT:
		// The first handshake is complete so move the transaction to the next state
		tx.State = pb.TransactionState_PENDING_ACKNOWLEDGED
	case pb.TransactionState_PENDING_ACKNOWLEDGED:
		// This is a complete transaction so update the database
		var account *db.Account
		if account, err = tx.GetAccount(s.parent.db); err != nil {
			log.Error().Err(err).Msg("could not fetch account from database")
			return fmt.Errorf("could not fetch account from database: %s", err)
		}

		account.Balance = account.Balance.Add(decimal.NewFromFloat(transaction.Amount))
		account.Completed++
		account.Pending--
		if err = s.parent.db.Save(&account).Error; err != nil {
			log.Error().Err(err).Msg("could not save beneficiary account")
			return fmt.Errorf("could not save beneficiary account: %s", err)
		}

		msg := fmt.Sprintf("ready for transaction %s: %.2f transfering from %s to %s", transaction.Txid, transaction.Amount, transaction.Originator, transaction.Beneficiary)
		s.parent.updates.Broadcast(0, msg, pb.MessageCategory_BLOCKCHAIN)
		tx.State = pb.TransactionState_COMPLETED
	default:
		log.Error().Str("state", tx.State.String()).Msg("unexpected transaction state")
		return fmt.Errorf("unexpected transaction state: %s", tx.State.String())
	}
	return nil
}

// sendRejected sends a rejected TRISA error message to the originator.
func (s *TRISA) sendRejected(tx *db.Transaction) (err error) {
	var (
		reject     *protocol.Error
		msg        *protocol.SecureEnvelope
		originator *db.Identity
	)

	// Fetch the originator address
	if originator, err = tx.GetOriginator(s.parent.db); err != nil {
		log.Error().Err(err).Msg("could not fetch originator address")
		return fmt.Errorf("could not fetch originator address")
	}

	// Conduct a GDS lookup if necessary to get the endpoint
	var peer *peers.Peer
	if peer, err = s.parent.peers.Search(originator.Provider); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return fmt.Errorf("could not search peer from directory service: %s", err)
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return status.Errorf(codes.FailedPrecondition, "could not exchange keys with remote peer: %s", err)
	}

	// Create the rejection message
	reject = protocol.Errorf(protocol.Rejected, "rejected by beneficiary")
	if msg, err = envelope.Reject(reject, envelope.WithEnvelopeID(tx.Envelope), envelope.WithRSAPublicKey(signKey)); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while creating reject envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transfer, handle errors
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return fmt.Errorf("could not perform TRISA exchange: %s", err)
	}

	// Check for the TRISA rejection error
	if state := envelope.Status(msg); state != envelope.Error {
		log.Error().Uint("state", uint(state)).Msg("unexpected TRISA response, expected error envelope")
		return fmt.Errorf("expected TRISA rejection error, received envelope in state %d", state)
	}

	tx.State = pb.TransactionState_FAILED

	return nil
}

// ConfirmAddress allows the rVASP to respond to proof-of-control requests.
func (s *TRISA) ConfirmAddress(ctx context.Context, in *protocol.Address) (out *protocol.AddressConfirmation, err error) {
	return nil, &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented address confirmation yet",
		Retry:   false,
	}
}

// KeyExchange facilitates signing key exchange between VASPs.
func (s *TRISA) KeyExchange(ctx context.Context, in *protocol.SigningKey) (out *protocol.SigningKey, err error) {
	var peer *peers.Peer
	if peer, err = s.parent.peers.FromContext(ctx); err != nil {
		log.Error().Err(err).Msg("could not verify peer from context")
		return nil, &protocol.Error{
			Code:    protocol.Unverified,
			Message: err.Error(),
		}
	}
	log.Info().Str("peer", peer.String()).Msg("key exchange request received")
	s.parent.updates.Broadcast(0, fmt.Sprintf("key exchange request received from %s", peer), pb.MessageCategory_TRISAP2P)

	// Cache key inside of the in-memory Peer map
	var pub interface{}
	if pub, err = x509.ParsePKIXPublicKey(in.Data); err != nil {
		log.Error().
			Err(err).
			Int64("version", in.Version).
			Str("algorithm", in.PublicKeyAlgorithm).
			Msg("could not parse incoming PKIX public key")
		return nil, protocol.Errorf(protocol.NoSigningKey, "could not parse signing key")
	}

	if err = peer.UpdateSigningKey(pub); err != nil {
		log.Error().Err(err).Msg("could not update signing key")
		return nil, protocol.Errorf(protocol.UnhandledAlgorithm, "unsuported signing algorithm")
	}

	// TODO: check not before and not after constraints

	// TODO: Kick off a go routine to store the key in the database

	// Return the public signing-key of the service
	// TODO: use separate signing key insead of using public key of mTLS certs
	var key *x509.Certificate
	if key, err = s.certs.GetLeafCertificate(); err != nil {
		log.Error().Err(err).Msg("could not extract leaf certificate")
		return nil, protocol.Errorf(protocol.InternalError, "could not return signing keys")
	}

	out = &protocol.SigningKey{
		Version:            int64(key.Version),
		Signature:          key.Signature,
		SignatureAlgorithm: key.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: key.PublicKeyAlgorithm.String(),
		NotBefore:          key.NotBefore.Format(time.RFC3339),
		NotAfter:           key.NotAfter.Format(time.RFC3339),
	}

	if out.Data, err = x509.MarshalPKIXPublicKey(key.PublicKey); err != nil {
		log.Error().Err(err).Msg("could not marshal PKIX public key")
		return nil, protocol.Errorf(protocol.InternalError, "could not marshal public key")
	}
	s.parent.updates.Broadcast(0, "keys marshaled, returning public keys for signing", pb.MessageCategory_TRISAP2P)

	return out, nil
}

// Status returns a directory health check status as online and requests half an hour checks.
func (s *TRISA) Status(ctx context.Context, in *protocol.HealthCheck) (out *protocol.ServiceState, err error) {
	log.Info().
		Uint32("attempts", in.Attempts).
		Str("last checked at", in.LastCheckedAt).
		Msg("status check")

	// Request another health check between 30 minutes and an hour from now.
	now := time.Now()
	return &protocol.ServiceState{
		Status:    protocol.ServiceState_HEALTHY,
		NotBefore: now.Add(30 * time.Minute).Format(time.RFC3339),
		NotAfter:  now.Add(1 * time.Hour).Format(time.RFC3339),
	}, nil
}

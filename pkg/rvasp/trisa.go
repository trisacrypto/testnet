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

	s.srv = grpc.NewServer(creds)
	protocol.RegisterTRISANetworkServer(s.srv, s)
	protocol.RegisterTRISAHealthServer(s.srv, s)

	var sock net.Listener
	if sock, err = net.Listen("tcp", s.parent.conf.TRISABindAddr); err != nil {
		return fmt.Errorf("trisa service could not listen on %q", s.parent.conf.TRISABindAddr)
	}

	go func() {
		defer sock.Close()

		log.Info().
			Str("listen", s.parent.conf.TRISABindAddr).
			Msg("trisa server started")

		if err := s.srv.Serve(sock); err != nil {
			s.parent.echan <- err
		}
	}()

	return nil
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
		log.Warn().Str("peer", peer.String()).Msg("no remote signing key available")
		s.parent.updates.Broadcast(0, "no remote signing key available, key exchange required", pb.MessageCategory_TRISAP2P)
		return nil, &protocol.Error{
			Code:    protocol.NoSigningKey,
			Message: "please retry transfer after key exchange",
			Retry:   true,
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
		log.Warn().Str("peer", peer.String()).Msg("no remote signing key available")
		s.parent.updates.Broadcast(0, "no remote signing key available, key exchange required", pb.MessageCategory_TRISAP2P)
		return &protocol.Error{
			Code:    protocol.NoSigningKey,
			Message: "please retry transfer after key exchange",
			Retry:   true,
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

	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unsupported identity type")
		return nil, protocol.Errorf(protocol.UnparseableIdentity, "rVASP requires ivms101.IdentityPayload payload identity type")
	}

	if payload.Transaction.TypeUrl != "type.googleapis.com/trisa.data.generic.v1beta1.Transaction" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported transaction type")
		return nil, protocol.Errorf(protocol.UnparseableTransaction, "rVASP requires trisa.data.generic.v1beta1.Transaction payload transaction type")
	}

	identity := &ivms101.IdentityPayload{}
	transaction := &generic.Transaction{}

	if err = payload.Identity.UnmarshalTo(identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity")
		return nil, protocol.Errorf(protocol.EnvelopeDecodeFail, "could not unmarshal identity")
	}
	if err = payload.Transaction.UnmarshalTo(transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal transaction")
		return nil, protocol.Errorf(protocol.EnvelopeDecodeFail, "could not unmarshal transaction")
	}
	s.parent.updates.Broadcast(0, fmt.Sprintf("secure envelope %s opened and payload decrypted and parsed", in.Id), pb.MessageCategory_TRISAP2P)

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

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending++
	if err = s.parent.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save beneficiary account")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// Update the identity with the beneficiary information.
	identity.BeneficiaryVasp = &ivms101.BeneficiaryVasp{}
	if identity.BeneficiaryVasp.BeneficiaryVasp, err = s.parent.vasp.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load beneficiary vasp")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	identity.Beneficiary = &ivms101.Beneficiary{
		BeneficiaryPersons: make([]*ivms101.Person, 0, 1),
		AccountNumbers:     []string{account.WalletAddress},
	}

	var beneficiary *ivms101.Person
	if beneficiary, err = account.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load beneficiary")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}
	identity.Beneficiary.BeneficiaryPersons = append(identity.Beneficiary.BeneficiaryPersons, beneficiary)

	// Update the transactionwith beneficiary information
	transaction.Beneficiary = account.WalletAddress
	transaction.Timestamp = time.Now().Format(time.RFC3339)

	// Fetch originator identity record
	var originatorIdentity db.Identity
	if err = s.parent.db.LookupIdentity(transaction.Originator).FirstOrInit(&originatorIdentity, db.Identity{}).Error; err != nil {
		log.Error().Err(err).Msg("could not lookup originator identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// If originator identity does not exist then create it
	if originatorIdentity.ID == 0 {
		originatorIdentity.WalletAddress = transaction.Originator
		originatorIdentity.Vasp = s.parent.vasp

		if err = s.parent.db.Create(&originatorIdentity).Error; err != nil {
			log.Error().Err(err).Msg("could not save originator identity")
			return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
		}
	}

	// Fetch beneficiary identity record
	var beneficiaryIdentity db.Identity
	if err = s.parent.db.LookupIdentity(transaction.Beneficiary).FirstOrInit(&beneficiaryIdentity, db.Identity{}).Error; err != nil {
		log.Error().Err(err).Msg("could not lookup identity")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	// If the beneficiary identity does not exist then create it
	if beneficiaryIdentity.ID == 0 {
		beneficiaryIdentity.WalletAddress = transaction.Beneficiary
		beneficiaryIdentity.Email = account.Email
		beneficiaryIdentity.Provider = s.parent.vasp.Name
		beneficiaryIdentity.Vasp = s.parent.vasp

		if err = s.parent.db.Create(&beneficiaryIdentity).Error; err != nil {
			log.Error().Err(err).Msg("could not save beneficiary identity")
			return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
		}
	}

	// Save the completed transaction in the database
	ach := db.Transaction{
		Envelope:    in.Id,
		Account:     account,
		Originator:  originatorIdentity,
		Beneficiary: beneficiaryIdentity,
		Amount:      decimal.NewFromFloat(transaction.Amount),
		Debit:       false,
		State:       db.TransactionCompleted,
		Timestamp:   time.Now(),
		Vasp:        s.parent.vasp,
	}

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
	account.Balance.Add(decimal.NewFromFloat(transaction.Amount))
	account.Completed++
	account.Pending--
	if err = s.parent.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save beneficiary account")
		return nil, protocol.Errorf(protocol.InternalError, "request could not be processed")
	}

	msg := fmt.Sprintf("ready for transaction %04d: %s transfering from %s to %s", ach.ID, ach.Amount, ach.Originator.WalletAddress, ach.Beneficiary.WalletAddress)
	s.parent.updates.Broadcast(0, msg, pb.MessageCategory_BLOCKCHAIN)

	// Encode and encrypt the payload information to return the secure envelope
	payload = &protocol.Payload{
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

	out, reject, err = envelope.Seal(payload, envelope.WithRSAPublicKey(peer.SigningKey()))
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

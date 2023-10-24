package rvasp

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	activity "github.com/trisacrypto/directory/pkg/utils/activity"
	"github.com/trisacrypto/directory/pkg/utils/logger"
	"github.com/trisacrypto/testnet/pkg"
	"github.com/trisacrypto/testnet/pkg/rvasp/config"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/envelope"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize zerolog with GCP logging requirements
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = logger.GCPFieldKeyTime
	zerolog.MessageFieldName = logger.GCPFieldKeyMsg

	// Add the severity hook for GCP logging
	var gcpHook logger.SeverityHook
	log.Logger = zerolog.New(os.Stdout).Hook(gcpHook).With().Timestamp().Logger()
}

// New creates a rVASP server with the specified configuration and prepares
// it to listen for and serve GRPC requests.
func New(conf *config.Config) (s *Server, err error) {
	if conf == nil {
		if conf, err = config.New(); err != nil {
			return nil, err
		}
	}

	// Set the global level
	zerolog.SetGlobalLevel(zerolog.Level(conf.LogLevel))

	// Set human readable logging if specified
	if conf.ConsoleLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	s = &Server{conf: conf, echan: make(chan error, 1)}
	if s.db, err = db.NewDB(conf); err != nil {
		return nil, err
	}
	s.vasp = s.db.GetVASP()

	// Create the TRISA service
	if s.trisa, err = NewTRISA(s); err != nil {
		return nil, fmt.Errorf("could not create TRISA service: %s", err)
	}

	// Create the remote peers using the same credentials as the TRISA service
	s.peers = peers.New(s.trisa.certs, s.trisa.chain, s.conf.GDS.URL)

	if s.conf.GDS.Insecure {
		// By default, the peers client connects via TLS. Making the explicit Connect()
		// call here without credentials will override that behavior and instead
		// connect insecurely for the purposes of local testing.
		s.peers.Connect(grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	s.updates = NewUpdateManager()

	// Start the activity publisher
	if err = activity.Start(conf.Activity); err != nil {
		return nil, fmt.Errorf("could not start the activity publisher: %s", err)
	}

	return s, nil
}

// Server implements the GRPC TRISAIntegration and TRISADemo services.
type Server struct {
	pb.UnimplementedTRISADemoServer
	pb.UnimplementedTRISAIntegrationServer
	conf    *config.Config
	srv     *grpc.Server
	db      *db.DB
	vasp    db.VASP
	trisa   *TRISA
	echan   chan error
	peers   *peers.Peers
	updates *UpdateManager
}

// Serve GRPC requests on the specified address.
func (s *Server) Serve() (err error) {
	// Initialize the gRPC server with panic recovery and tracing
	s.srv = grpc.NewServer(grpc.UnaryInterceptor(UnaryTraceInterceptor), grpc.StreamInterceptor(StreamTraceInterceptor))
	pb.RegisterTRISADemoServer(s.srv, s)
	pb.RegisterTRISAIntegrationServer(s.srv, s)

	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		s.echan <- s.Shutdown()
	}()

	// Run the TRISA service on the TRISABindAddr
	if err = s.trisa.Serve(); err != nil {
		return err
	}

	// Listen for TCP requests on the specified address and port
	var sock net.Listener
	if sock, err = net.Listen("tcp", s.conf.BindAddr); err != nil {
		return fmt.Errorf("could not listen on %q", s.conf.BindAddr)
	}
	defer sock.Close()

	// Run the server
	go func() {
		log.Info().
			Str("listen", s.conf.BindAddr).
			Str("version", pkg.Version()).
			Str("name", s.vasp.Name).
			Msg("server started")

		if err := s.srv.Serve(sock); err != nil {
			s.echan <- err
		}
	}()

	// Listen for any errors that might have occurred and wait for all go routines to finish
	if err = <-s.echan; err != nil {
		return err
	}
	return nil
}

// Shutdown the rVASP Service gracefully
func (s *Server) Shutdown() (err error) {
	log.Info().Msg("gracefully shutting down")
	s.srv.GracefulStop()
	if err = s.trisa.Shutdown(); err != nil {
		log.Error().Err(err).Msg("could not shutdown trisa server")
		return err
	}
	log.Debug().Msg("successful shutdown")
	return nil
}

// Transfer accepts a transfer request from a beneficiary and begins the InterVASP
// protocol to perform identity verification prior to establishing the transaction in
// the blockchain between crypto wallet addresses.
func (s *Server) Transfer(ctx context.Context, req *pb.TransferRequest) (reply *pb.TransferReply, err error) {
	// Get originator account and confirm it belongs to this RVASP
	var account db.Account
	if err = s.db.LookupAccount(req.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("account", req.Account).Msg("not found")
			return nil, status.Error(codes.NotFound, "account not found")
		}
		log.Error().Err(err).Msg("could not lookup account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup account: %s", err)
	}

	// Retrieve the policy for the originator account
	var wallet db.Wallet
	if err = s.db.LookupWallet(account.WalletAddress).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("wallet", account.WalletAddress).Msg("not found")
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		log.Error().Err(err).Msg("could not lookup wallet")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup wallet: %s", err)
	}

	// Fetch the beneficiary Wallet
	var beneficiary *db.Wallet
	if beneficiary, err = s.fetchBeneficiaryWallet(req); err != nil {
		return nil, err
	}

	// Create a new Transaction
	var xfer *db.Transaction
	if xfer, err = s.db.MakeTransaction(account.WalletAddress, beneficiary.Address); err != nil {
		return nil, err
	}
	xfer.Account = account
	xfer.Amount = decimal.NewFromFloat32(req.Amount)
	xfer.AssetType = req.AssetType
	xfer.Debit = true

	// Run the scenario for the wallet's configured policy
	var transferError error
	policy := wallet.OriginatorPolicy
	log.Debug().Str("wallet", account.WalletAddress).Str("policy", string(policy)).Msg("initiating transfer")
	switch policy {
	case db.SendPartial:
		// Send a transfer request to the beneficiary containing partial beneficiary
		// identity information.
		transferError = s.sendTransfer(xfer, beneficiary, true)
	case db.SendFull:
		// Send a transfer request to the beneficiary containing full beneficiary
		// identity information.
		transferError = s.sendTransfer(xfer, beneficiary, false)
	case db.SendError:
		// Send a TRISA error to the beneficiary.
		transferError = s.sendError(xfer, beneficiary)
	default:
		log.Error().Str("wallet", account.WalletAddress).Str("policy", string(policy)).Msg("unknown policy")
		return nil, status.Errorf(codes.FailedPrecondition, "unknown originator policy '%s' for wallet '%s'", policy, account.WalletAddress)
	}

	// Build the transfer response
	reply = &pb.TransferReply{}

	// Handle rVASP errors and TRISA protocol errors
	if transferError != nil {
		switch err := transferError.(type) {
		case *protocol.Error:
			log.Warn().Str("message", err.Error()).Msg("TRISA protocol error while performing transfer")
			reply.Error = &pb.Error{
				Code:    int32(err.Code),
				Message: err.Message,
			}
			xfer.SetState(pb.TransactionState_REJECTED)
			transferError = nil
		default:
			log.Warn().Err(err).Msg("error while performing transfer")
			xfer.SetState(pb.TransactionState_FAILED)
		}
	}

	// Populate the transfer response with the transaction details
	reply.Transaction = xfer.Proto()

	// Save the updated transaction
	// TODO: Clean up completed transactions in the database
	if err = s.db.Save(xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction")
		return nil, status.Errorf(codes.Internal, "could not save transaction: %s", err)
	}

	return reply, transferError
}

// fetchBeneficiary fetches the beneficiary Wallet from the request.
func (s *Server) fetchBeneficiaryWallet(req *pb.TransferRequest) (wallet *db.Wallet, err error) {
	if req.BeneficiaryVasp != "" {
		// If a beneficiary VASP is provided, assume the transfer is to an external
		// VASP (not a local wallet)
		wallet = &db.Wallet{
			Address: req.Beneficiary,
			Provider: db.VASP{
				Name: req.BeneficiaryVasp,
			},
			Vasp: s.vasp,
		}
	} else {
		// Otherwise attempt to lookup the beneficiary wallet from the database
		wallet = &db.Wallet{}
		if err = s.db.LookupAnyBeneficiary(req.Beneficiary).First(wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Info().Str("beneficiary", req.Beneficiary).Msg("not found")
				return nil, status.Error(codes.NotFound, "beneficiary not found (use external_demo?)")
			}
			log.Error().Err(err).Msg("could not lookup beneficiary")
			return nil, status.Errorf(codes.FailedPrecondition, "could not lookup beneficiary: %s", err)
		}

		if req.CheckBeneficiary {
			if req.BeneficiaryVasp != wallet.Provider.Name {
				log.Warn().
					Str("expected", req.BeneficiaryVasp).
					Str("actual", wallet.Provider.Name).
					Msg("check beneficiary failed")
				return nil, status.Error(codes.InvalidArgument, "beneficiary wallet does not match beneficiary VASP")
			}
		}
	}
	return wallet, nil
}

// sendTransfer looks up the beneficiary from the request and sends a transfer request
// to the beneficiary. If partial is true, then the full beneficiary identity
// information is not included in the payload. This function handles pending responses
// from the beneficiary saving the transaction in an "await" state in the database.
func (s *Server) sendTransfer(xfer *db.Transaction, beneficiary *db.Wallet, partial bool) (err error) {
	// Fetch the remote peer
	var peer *peers.Peer
	if peer, err = s.fetchPeer(beneficiary.Provider.Name); err != nil {
		log.Warn().Err(err).Msg("could not fetch beneficiary peer")
		return status.Errorf(codes.FailedPrecondition, "could not fetch beneficiary peer: %s", err)
	}

	// Fetch the signing key
	var signKey *rsa.PublicKey
	if signKey, err = s.fetchSigningKey(peer); err != nil {
		log.Warn().Err(err).Msg("could not fetch signing key from beneficiary peer")
		return status.Errorf(codes.FailedPrecondition, "could not fetch signing key from beneficiary peer: %s", err)
	}

	if err = s.db.Create(xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save pending transaction")
		return status.Errorf(codes.FailedPrecondition, "could not save pending transaction: %s", err)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	xfer.Account.Pending++
	if err = s.db.Save(&xfer.Account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return status.Errorf(codes.FailedPrecondition, "could not save originator account: %s", err)
	}

	// Create an identity and transaction payload for TRISA exchange
	transaction := &generic.Transaction{
		Originator:  xfer.Account.WalletAddress,
		Beneficiary: beneficiary.Address,
		Network:     "TestNet",
		AssetType:   xfer.AssetType,
		Timestamp:   xfer.Timestamp.Format(time.RFC3339),
	}

	// Set the amount on the transaction payload
	transaction.Amount, _ = xfer.Amount.Float64()

	var beneficiaryAccount db.Account
	if partial {
		// If partial is specified then only populate the beneficiary address
		beneficiaryAccount = db.Account{
			WalletAddress: beneficiary.Address,
		}
	} else {
		// If partial is false then retrieve the full beneficiary account
		if err = s.db.LookupAnyAccount(beneficiary.Address).First(&beneficiaryAccount).Error; err != nil {
			log.Warn().Err(err).Msg("could not lookup remote beneficiary account")
			return status.Errorf(codes.FailedPrecondition, "could not lookup remote beneficiary account: %s", err)
		}
	}

	var identity *ivms101.IdentityPayload
	if identity, err = s.createIdentityPayload(xfer.Account, beneficiaryAccount); err != nil {
		return err
	}

	var payload *protocol.Payload
	if payload, err = createTransferPayload(identity, transaction); err != nil {
		log.Error().Err(err).Msg("could not create transfer payload")
		return status.Errorf(codes.Internal, "could not create transfer payload: %s", err)
	}

	// Secure the envelope with the remote beneficiary's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(xfer.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Warn().Err(err).Msg("TRISA protocol error while sealing envelope")
		return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Warn().Err(err).Msg("could not perform TRISA exchange")
		return status.Errorf(codes.FailedPrecondition, "could not perform TRISA exchange: %s", err)
	}

	// Check for TRISA rejection errors
	reject, isErr := envelope.Check(msg)
	if isErr {
		if reject != nil {
			return reject
		}
		log.Warn().Err(err).Msg("TRISA protocol error while checking envelope")
		return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.trisa.sign))
	if err != nil {
		log.Warn().Err(err).Msg("TRISA protocol error while opening envelope")
		return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Parse the response payload
	var pending *generic.Pending
	var parseError *protocol.Error
	if identity, transaction, pending, parseError = parsePayload(payload, true); parseError != nil {
		log.Warn().Str("message", parseError.Message).Msg("TRISA protocol error while parsing payload")
		return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", parseError.Message)
	}

	// Update the transaction record with the identity payload
	var data []byte
	if data, err = protojson.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal IVMS 101 identity")
		return status.Errorf(codes.Internal, "could not marshal IVMS 101 identity: %s", err)
	}
	xfer.Identity = string(data)

	// Handle both synchronous and asynchronous responses from the beneficiary
	if pending != nil {
		// Update the Transaction in the database with the pending timestamps
		if xfer.NotBefore, err = time.Parse(time.RFC3339, pending.ReplyNotBefore); err != nil {
			log.Warn().Err(err).Msg("TRISA protocol error: could not parse ReplyNotBefore timestamp")
			return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse ReplyNotBefore timestamp in pending message: %s", err)
		}

		if xfer.NotAfter, err = time.Parse(time.RFC3339, pending.ReplyNotAfter); err != nil {
			log.Warn().Err(err).Msg("TRISA protocol error: could not parse ReplyNotAfter timestamp")
			return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse ReplyNotAfter timestamp in pending message: %s", err)
		}
	} else if transaction != nil {
		if !partial {
			// Validate that the beneficiary identity matches the original request
			if identity.BeneficiaryVasp == nil || identity.BeneficiaryVasp.BeneficiaryVasp == nil {
				log.Warn().Msg("TRISA protocol error: missing beneficiary vasp identity")
				return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary vasp identity")
			}

			beneficiaryInfo := identity.Beneficiary
			if beneficiaryInfo == nil {
				log.Warn().Msg("TRISA protocol error: missing beneficiary identity")
				return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary identity")
			}

			if len(beneficiaryInfo.BeneficiaryPersons) != 1 {
				log.Warn().Int("persons", len(beneficiaryInfo.BeneficiaryPersons)).Msg("TRISA protocol error: unexpected number of beneficiary persons")
				return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: unexpected number of beneficiary persons")
			}

			if len(beneficiaryInfo.AccountNumbers) != 1 {
				log.Warn().Int("accounts", len(beneficiaryInfo.AccountNumbers)).Msg("TRISA protocol error: unexpected number of beneficiary accounts")
				return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: unexpected number of beneficiary accounts")
			}

			// TODO: Validate that the actual address was returned
			if beneficiaryInfo.AccountNumbers[0] == "" {
				log.Warn().Msg("TRISA protocol error: missing beneficiary address")
				return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary address")
			}
		}

		// Update the account information
		xfer.Account.Pending--
		xfer.Account.Completed++
		xfer.Account.Balance = xfer.Account.Balance.Sub(xfer.Amount)
		if err = s.db.Save(xfer.Account).Error; err != nil {
			log.Error().Err(err).Msg("could not save originator account")
			return status.Errorf(codes.Internal, "could not save originator account: %s", err)
		}

		// This transaction is now complete
		xfer.SetState(pb.TransactionState_COMPLETED)
		xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)
	}

	return nil
}

// sendError sends a TRISA error to the beneficiary.
func (s *Server) sendError(xfer *db.Transaction, beneficiary *db.Wallet) (err error) {
	// Fetch the remote peer
	var peer *peers.Peer
	if peer, err = s.fetchPeer(beneficiary.Provider.Name); err != nil {
		log.Warn().Err(err).Msg("could not fetch beneficiary peer")
		return status.Errorf(codes.FailedPrecondition, "could not fetch beneficiary peer: %s", err)
	}

	reject := protocol.Errorf(protocol.ComplianceCheckFail, "rVASP mock compliance check failed")
	var msg *protocol.SecureEnvelope
	if msg, err = envelope.Reject(reject, envelope.WithEnvelopeID(xfer.Envelope)); err != nil {
		log.Error().Err(err).Msg("could not create TRISA error envelope")
		return status.Errorf(codes.Internal, "could not create TRISA error envelope: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Warn().Err(err).Msg("could not perform TRISA exchange")
		return status.Errorf(codes.FailedPrecondition, "could not perform TRISA exchange: %s", err)
	}

	// Check for the TRISA rejection error
	reject, isErr := envelope.Check(msg)
	if !isErr || reject == nil {
		state := envelope.Status(msg)
		log.Warn().Str("state", state.String()).Msg("unexpected TRISA response, expected reject envelope")
		return fmt.Errorf("expected TRISA rejection error, received envelope in state %s", state.String())
	}
	xfer.SetState(pb.TransactionState_REJECTED)

	return reject
}

// respondAsync responds to a serviced transfer request from the beneficiary by
// continuing or completing the asynchronous handshake.
func (s *Server) respondAsync(peer *peers.Peer, payload *protocol.Payload, identity *ivms101.IdentityPayload, transaction *generic.Transaction, xfer *db.Transaction) (out *protocol.SecureEnvelope, transferError *protocol.Error) {
	// Secure envelope was successfully received
	now := time.Now()

	// Verify that the transaction has not expired
	if now.Before(xfer.NotBefore) || now.After(xfer.NotAfter) {
		log.Debug().Time("now", now).Time("not_before", xfer.NotBefore).Time("not_after", xfer.NotAfter).Str("id", xfer.Envelope).Msg("received expired async transaction")
		return nil, protocol.Errorf(protocol.ComplianceCheckFail, "received expired transaction")
	}

	// Fetch the signing key from the remote peer
	var signKey *rsa.PublicKey
	var err error
	if signKey, err = s.fetchSigningKey(peer); err != nil {
		log.Warn().Err(err).Msg("could not fetch signing key from beneficiary peer")
		return nil, protocol.Errorf(protocol.NoSigningKey, "could not fetch signing key from beneficiary peer: %s", err)
	}

	// Marshal the identity payload into the transaction record
	var data []byte
	if data, err = protojson.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal identity payload")
		return nil, protocol.Errorf(protocol.InternalError, "could not marshal identity payload: %s", err)
	}
	xfer.Identity = string(data)

	// Marshal the transaction payload into the transaction record
	if data, err = protojson.Marshal(transaction); err != nil {
		log.Error().Err(err).Msg("could not marshal transaction payload")
		return nil, protocol.Errorf(protocol.InternalError, "could not marshal transaction payload: %s", err)
	}
	xfer.Transaction = string(data)

	// Fetch the beneficiary identity
	if err = s.db.LookupIdentity(transaction.Beneficiary).First(&xfer.Beneficiary).Error; err != nil {
		log.Error().Err(err).Str("wallet_address", transaction.Beneficiary).Msg("could not lookup beneficiary identity")
		return nil, protocol.Errorf(protocol.InternalError, "could not lookup beneficiary identity: %s", err)
	}

	// Save the peer name so we can access it later
	xfer.Beneficiary.Provider = peer.String()
	if err = s.db.Save(&xfer.Beneficiary).Error; err != nil {
		log.Error().Err(err).Msg("could not save beneficiary identity")
		return nil, protocol.Errorf(protocol.InternalError, "could not save beneficiary identity: %s", err)
	}

	// Create the response envelope
	out, reject, err := envelope.Seal(payload, envelope.WithRSAPublicKey(signKey), envelope.WithEnvelopeID(xfer.Envelope))
	if err != nil {
		if reject != nil {
			if out, err = envelope.Reject(reject, envelope.WithEnvelopeID(xfer.Envelope)); err != nil {
				return nil, protocol.Errorf(protocol.EnvelopeDecodeFail, "TRISA protocol error: %s", err)
			}
			xfer.SetState(pb.TransactionState_REJECTED)
			return out, nil
		}
		log.Warn().Err(err).Msg("TRISA protocol error while sealing envelope")
		return nil, protocol.Errorf(protocol.EnvelopeDecodeFail, "TRISA protocol error: %s", err)
	}

	// Check the transaction state
	switch xfer.State {
	case pb.TransactionState_AWAITING_REPLY:
		// Mark the transaction as pending for the async routine
		xfer.SetState(pb.TransactionState_PENDING_RECEIVED)
	case pb.TransactionState_ACCEPTED:
		// The handshake is complete, finalize the transaction
		var account db.Account
		if err = s.db.LookupAccount(transaction.Originator).First(&account).Error; err != nil {
			log.Warn().Err(err).Msg("could not find originator account")
			return nil, protocol.Errorf(protocol.InternalError, "could not find originator account: %s", err)
		}

		// Save the pending transaction on the account
		// TODO: remove pending transactions
		account.Pending--
		account.Completed++
		account.Balance = account.Balance.Sub(xfer.Amount)
		if err = s.db.Save(&account).Error; err != nil {
			log.Error().Err(err).Msg("could not save originator account")
			return nil, protocol.Errorf(protocol.InternalError, "could not save originator account: %s", err)
		}
		xfer.SetState(pb.TransactionState_COMPLETED)
	default:
		log.Error().Str("state", xfer.State.String()).Msg("unexpected transaction state")
		return nil, protocol.Errorf(protocol.ComplianceCheckFail, "unexpected transaction state: %s", xfer.State.String())
	}

	return out, nil
}

// continueAsync continues an asynchronous transaction by sending a new transfer to the
// beneficiary with a populated TxID.
func (s *Server) continueAsync(xfer *db.Transaction) (err error) {
	// Unmarshal the transaction payload from the transaction record
	transaction := &generic.Transaction{}
	if err = protojson.Unmarshal([]byte(xfer.Transaction), transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal transaction payload")
		return fmt.Errorf("could not unmarshal transaction payload: %s", err)
	}

	// Unmarshal the identity payload from the identity record
	identity := &ivms101.IdentityPayload{}
	if err = protojson.Unmarshal([]byte(xfer.Identity), identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity payload")
		return fmt.Errorf("could not unmarshal identity payload: %s", err)
	}

	// Fetch the beneficiary name from the database
	var beneficiary *db.Identity
	if beneficiary, err = xfer.GetBeneficiary(s.db); err != nil {
		log.Error().Err(err).Msg("could not fetch beneficiary address")
		return fmt.Errorf("could not fetch beneficiary address")
	}

	// Fetch the remote peer
	var peer *peers.Peer
	if peer, err = s.fetchPeer(beneficiary.Provider); err != nil {
		log.Error().Err(err).Msg("could not fetch beneficiary peer")
		return fmt.Errorf("could not fetch beneficiary peer: %s", err)
	}

	// Fetch the signing key from the remote peer
	var signKey *rsa.PublicKey
	if signKey, err = s.fetchSigningKey(peer); err != nil {
		log.Warn().Err(err).Msg("could not fetch signing key from beneficiary peer")
		return fmt.Errorf("could not fetch signing key from beneficiary peer: %s", err)
	}

	// Fill the transaction with a new TxID to continue the handshake
	var payload *protocol.Payload
	transaction.Txid = uuid.New().String()
	if payload, err = createTransferPayload(identity, transaction); err != nil {
		log.Error().Err(err).Msg("could not create transfer payload")
		return fmt.Errorf("could not create transfer payload: %s", err)
	}

	// Secure the envelope with the remote beneficiary's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(xfer.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Warn().Err(err).Msg("TRISA protocol error while sealing envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Warn().Err(err).Msg("could not perform TRISA exchange")
		return fmt.Errorf("could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.trisa.sign))
	if err != nil {
		log.Warn().Err(err).Msg("TRISA protocol error while opening envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	// Parse the response payload
	var pending *generic.Pending
	var parseError *protocol.Error
	if identity, _, pending, parseError = parsePayload(payload, true); parseError != nil {
		log.Warn().Str("message", parseError.Message).Msg("TRISA protocol error while parsing payload")
		return fmt.Errorf("TRISA protocol error: %s", parseError.Message)
	}

	if pending == nil {
		log.Warn().Msg("TRISA protocol error: expected pending response")
		return fmt.Errorf("TRISA protocol error: expected pending response")
	}

	// Update the transaction with the identity information
	var data []byte
	if data, err = protojson.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal IVMS 101 identity")
		return fmt.Errorf("could not marshal IVMS 101 identity: %s", err)
	}
	xfer.Identity = string(data)

	// Update the transaction with the new generic.Transaction
	if data, err = protojson.Marshal(transaction); err != nil {
		log.Error().Err(err).Msg("could not marshal generic.Transaction")
		return fmt.Errorf("could not marshal generic.Transaction: %s", err)
	}
	xfer.Transaction = string(data)

	// Update the Transaction in the database with the pending timestamps
	if xfer.NotBefore, err = time.Parse(time.RFC3339, pending.ReplyNotBefore); err != nil {
		log.Warn().Err(err).Msg("TRISA protocol error: could not parse ReplyNotBefore timestamp")
		return fmt.Errorf("TRISA protocol error: could not parse ReplyNotBefore timestamp: %s", err)
	}

	if xfer.NotAfter, err = time.Parse(time.RFC3339, pending.ReplyNotAfter); err != nil {
		log.Warn().Err(err).Msg("TRISA protocol error: could not parse ReplyNotAfter timestamp")
		return fmt.Errorf("TRISA protocol error: could not parse ReplyNotAfter timestamp: %s", err)
	}

	xfer.SetState(pb.TransactionState_ACCEPTED)
	return nil
}

// AccountStatus is a demo RPC to allow demo clients to fetch their recent transactions.
func (s *Server) AccountStatus(ctx context.Context, req *pb.AccountRequest) (rep *pb.AccountReply, err error) {
	rep = &pb.AccountReply{}

	// Lookup the account in the database
	var account db.Account
	if err = s.db.LookupAccount(req.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Err(err).Msg("account not found")
			return nil, status.Error(codes.NotFound, "account not found")
		}
		log.Warn().Err(err).Msg("could not lookup account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup account: %s", err)
	}

	rep.Name = account.Name
	rep.Email = account.Email
	rep.WalletAddress = account.WalletAddress
	rep.Balance = account.BalanceFloat()
	rep.Completed = account.Completed
	rep.Pending = account.Pending

	if !req.NoTransactions {
		var transactions []db.Transaction
		if transactions, err = account.Transactions(s.db); err != nil {
			log.Error().Err(err).Msg("could not get transactions")
			return nil, status.Errorf(codes.FailedPrecondition, "could not get transactions: %s", err)
		}

		rep.Transactions = make([]*pb.Transaction, 0, len(transactions))
		for _, transaction := range transactions {
			rep.Transactions = append(rep.Transactions, transaction.Proto())
		}
	}

	log.Info().
		Str("account", rep.Email).
		Int("transactions", len(rep.Transactions)).
		Msg("account status")
	return rep, nil
}

// LiveUpdates is a demo bidirectional RPC that allows demo clients to explicitly show
// the message interchange between VASPs during the InterVASP protocol. The demo client
// connects to both sides of a transaction and can push commands to the stream; any
// messages received by the VASP as they perform the protocol are sent down to the UI.
func (s *Server) LiveUpdates(stream pb.TRISADemo_LiveUpdatesServer) (err error) {
	var (
		client   string
		messages uint64
	)

	ctx := stream.Context()

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req *pb.Command
		if req, err = stream.Recv(); err != nil {
			// The stream was closed on the client side
			if err == io.EOF {
				if client == "" {
					log.Warn().Msg("live updates connection closed before first message")
				} else {
					log.Warn().Str("client", client).Msg("live updates connection closed")
				}
			}

			// Some other error occurred
			log.Error().Err(err).Str("client", client).Msg("connection dropped")
			return nil
		}

		// If this is the first time we've seen the client, log it
		if client == "" {
			client = req.Client
			if err = s.updates.Add(client, stream); err != nil {
				log.Error().Err(err).Msg("could not create client updater")
				return err
			}
			log.Info().Str("client", client).Msg("connected to live updates")
			defer s.updates.Del(client)

		} else if client != req.Client {
			log.Warn().Str("request from", req.Client).Str("client stream", client).Msg("unexpected client")
			s.updates.Del(client)
			return fmt.Errorf("unexpected client %q (connected as %q)", req.Client, client)
		}

		// Handle the message
		messages++
		log.Info().
			Uint64("message", messages).
			Str("type", req.Type.String()).
			Msg("received message")

		switch req.Type {
		case pb.RPC_NORPC:
			// Send back an acknowledgement message
			ack := &pb.Message{
				Type:      pb.RPC_NORPC,
				Id:        req.Id,
				Update:    fmt.Sprintf("command %d acknowledged", req.Id),
				Timestamp: time.Now().Format(time.RFC3339),
			}
			if err = s.updates.Send(client, ack); err != nil {
				log.Error().Err(err).Str("client", client).Msg("could not send message")
				return err
			}
		case pb.RPC_ACCOUNT:
			var rep *pb.AccountReply
			if rep, err = s.AccountStatus(context.Background(), req.GetAccount()); err != nil {
				return err
			}

			ack := &pb.Message{
				Type:      pb.RPC_ACCOUNT,
				Id:        req.Id,
				Timestamp: time.Now().Format(time.RFC3339),
				Reply:     &pb.Message_Account{Account: rep},
			}

			if err = s.updates.Send(client, ack); err != nil {
				log.Error().Err(err).Str("client", client).Msg("could not send message")
				return err
			}
		case pb.RPC_TRANSFER:
			if err = s.handleTransaction(client, req); err != nil {
				log.Error().Err(err).Msg("could not handle transaction")
				return err
			}
		}
	}
}

// NOTE: this adds in some purposeful latency to make the demo easier to see
func (s *Server) handleTransaction(client string, req *pb.Command) (err error) {
	// Get the transfer from the original command, will panic if nil
	transfer := req.GetTransfer()
	message := fmt.Sprintf("starting transaction of %0.2f from %s to %s", transfer.Amount, transfer.Account, transfer.Beneficiary)
	s.updates.Broadcast(req.Id, message, pb.MessageCategory_LEDGER)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Handle Demo UI errors before the account lookup
	if transfer.OriginatingVasp != "" && transfer.OriginatingVasp != s.vasp.Name {
		log.Info().Str("requested", transfer.OriginatingVasp).Str("local", s.vasp.Name).Msg("requested originator does not match local VASP")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrWrongVASP, "message sent to the wrong originator VASP"),
		)
	}

	// Lookup the account associated with the transfer originator
	var account db.Account
	if err = s.db.LookupAccount(transfer.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("account", transfer.Account).Msg("not found")
			return s.updates.SendTransferError(client, req.Id,
				pb.Errorf(pb.ErrNotFound, "account not found"),
			)
		}
		return fmt.Errorf("could not fetch account: %s", err)
	}
	s.updates.Broadcast(req.Id, fmt.Sprintf("account %04d accessed successfully", account.ID), pb.MessageCategory_LEDGER)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Lookup the wallet of the beneficiary
	var beneficiary db.Wallet
	if err = s.db.LookupBeneficiary(transfer.Beneficiary).First(&beneficiary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("beneficiary", transfer.Beneficiary).Msg("not found")
			return s.updates.SendTransferError(client, req.Id,
				pb.Errorf(pb.ErrNotFound, "beneficiary wallet not found"),
			)
		}
		return fmt.Errorf("could not fetch beneficiary wallet: %s", err)
	}

	if transfer.CheckBeneficiary {
		if transfer.BeneficiaryVasp != beneficiary.Provider.Name {
			log.Info().
				Str("expected", transfer.BeneficiaryVasp).
				Str("actual", beneficiary.Provider.Name).
				Msg("check beneficiary failed")
			return s.updates.SendTransferError(client, req.Id,
				pb.Errorf(pb.ErrWrongVASP, "beneficiary wallet does not match beneficiary vasp"),
			)
		}
	}
	s.updates.Broadcast(req.Id, fmt.Sprintf("wallet %s provided by %s", beneficiary.Address, beneficiary.Provider.Name), pb.MessageCategory_BLOCKCHAIN)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// TODO: lookup peer from cache rather than always doing a directory service lookup
	var peer *peers.Peer
	s.updates.Broadcast(req.Id, fmt.Sprintf("search for %s in directory service", beneficiary.Provider.Name), pb.MessageCategory_TRISADS)
	// send search request activity to network activity handler
	activity.Search().Add()
	if peer, err = s.peers.Search(beneficiary.Provider.Name); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not search peer from directory service"),
		)
	}
	info := peer.Info()
	s.updates.Broadcast(req.Id, fmt.Sprintf("identified TRISA remote peer %s at %s via directory service", info.ID, info.Endpoint), pb.MessageCategory_TRISADS)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	var signKey *rsa.PublicKey
	s.updates.Broadcast(req.Id, "exchanging peer signing keys", pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not exchange keys with remote peer"),
		)
	}

	// Prepare the transaction
	// Save the pending transaction and increment the accounts pending field
	xfer := db.Transaction{
		Envelope: uuid.New().String(),
		Account:  account,
		Amount:   decimal.NewFromFloat32(transfer.Amount),
		Debit:    true,
		State:    pb.TransactionState_AWAITING_REPLY,
		Vasp:     s.vasp,
	}

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save pending transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not save pending transaction"),
		)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending++
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not save originator account"),
		)
	}

	s.updates.Broadcast(req.Id, "ready to execute transaction", pb.MessageCategory_BLOCKCHAIN)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Create an identity and transaction payload for TRISA exchange
	transaction := &generic.Transaction{
		Txid:        fmt.Sprintf("%d", xfer.ID),
		Originator:  account.WalletAddress,
		Beneficiary: beneficiary.Address,
		Amount:      float64(transfer.Amount),
		Network:     "TestNet",
		Timestamp:   xfer.Timestamp.Format(time.RFC3339),
	}
	identity := &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
	}
	if identity.OriginatingVasp.OriginatingVasp, err = s.vasp.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load originator vasp")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not load originator vasp"),
		)
	}

	identity.Originator = &ivms101.Originator{
		OriginatorPersons: make([]*ivms101.Person, 0, 1),
		AccountNumbers:    []string{account.WalletAddress},
	}
	var originator *ivms101.Person
	if originator, err = account.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load originator identity")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not load originator identity"),
		)
	}
	identity.Originator.OriginatorPersons = append(identity.Originator.OriginatorPersons, originator)

	payload := &protocol.Payload{}
	if payload.Transaction, err = anypb.New(transaction); err != nil {
		log.Error().Err(err).Msg("could not serialize transaction payload")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not serialize transaction payload"),
		)
	}
	if payload.Identity, err = anypb.New(identity); err != nil {
		log.Error().Err(err).Msg("could not serialize identity payload")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not serialize identity payload"),
		)
	}

	s.updates.Broadcast(req.Id, "transaction and identity payload constructed", pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Secure the envelope with the remote beneficiary's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(xfer.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	s.updates.Broadcast(req.Id, fmt.Sprintf("secure envelope %s sealed: encrypted with AES-GCM and RSA - sending ...", msg.Id), pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	s.updates.Broadcast(req.Id, fmt.Sprintf("received %s information exchange reply from %s", msg.Id, peer.String()), pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.trisa.sign))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Verify the contents of the response
	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unsupported identity type")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "unsupported identity type", payload.Identity.TypeUrl),
		)
	}

	if payload.Transaction.TypeUrl != "type.googleapis.com/trisa.data.generic.v1beta1.Transaction" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported transaction type")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "unsupported transaction type", payload.Transaction.TypeUrl),
		)
	}

	identity = &ivms101.IdentityPayload{}
	transaction = &generic.Transaction{}
	if err = payload.Identity.UnmarshalTo(identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}
	if err = payload.Transaction.UnmarshalTo(transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	s.updates.Broadcast(req.Id, "successfully decrypted and parsed secure envelope", pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Update the completed transaction and save to disk
	xfer.Beneficiary = db.Identity{
		WalletAddress: transaction.Beneficiary,
		Vasp:          s.vasp,
	}
	xfer.SetState(pb.TransactionState_COMPLETED)
	xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)

	// Serialize the identity information as JSON data
	var data []byte
	if data, err = json.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not save completed transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not marshal IVMS 101 identity"),
		)
	}
	xfer.Identity = string(data)

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save completed transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending--
	account.Completed++
	account.Balance = account.Balance.Sub(xfer.Amount)
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction on account")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	message = fmt.Sprintf("transaction %04d complete: %s transferred from %s to %s", xfer.ID, xfer.Amount.String(), xfer.Originator.WalletAddress, xfer.Beneficiary.WalletAddress)
	s.updates.Broadcast(req.Id, message, pb.MessageCategory_BLOCKCHAIN)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	s.updates.Broadcast(req.Id, fmt.Sprintf("%04d new account balance: %s", account.ID, account.Balance), pb.MessageCategory_LEDGER)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	rep := &pb.Message{
		Type:      pb.RPC_TRANSFER,
		Id:        req.Id,
		Timestamp: time.Now().Format(time.RFC3339),
		Category:  pb.MessageCategory_LEDGER,
		Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
			Transaction: xfer.Proto(),
		}},
	}

	return s.updates.Send(client, rep)
}

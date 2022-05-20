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
	s.peers.Connect(grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.updates = NewUpdateManager()
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
	// Initialize the gRPC server
	s.srv = grpc.NewServer()
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
func (s *Server) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferReply, error) {
	// Get originator account and confirm it belongs to this RVASP
	var account db.Account
	if err := s.db.LookupAccount(req.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("account", req.Account).Msg("not found")
			return nil, status.Error(codes.NotFound, "account not found")
		}
		log.Error().Err(err).Msg("could not lookup account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup account: %s", err)
	}

	// Retrieve the policy for the originator account
	var wallet db.Wallet
	if err := s.db.LookupWallet(account.WalletAddress).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("wallet", account.WalletAddress).Msg("not found")
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		log.Error().Err(err).Msg("could not lookup wallet")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup wallet: %s", err)
	}

	// Run the scenario for the wallet's configured policy
	policy := wallet.Policy
	switch policy {
	case db.BasicSync:
		return s.basicSyncTransfer(req, account)
	case db.PartialSync:
		return s.partialSyncTransfer(req, account)
	case db.FullAsync:
		return s.AsyncTransfer(req, account)
	case db.RejectedAsync:
		return s.AsyncTransfer(req, account)
	default:
		return nil, status.Errorf(codes.FailedPrecondition, "unknown policy '%s' for wallet '%s'", policy, account.Wallet.Address)
	}
}

// fetchBeneficiary fetches the beneficiary Wallet from the request.
func (s *Server) fetchBeneficiaryWallet(req *pb.TransferRequest) (wallet *db.Wallet, err error) {
	if req.ExternalDemo {
		// If external demo is enabled, then create the Wallet from the request parameters
		if req.BeneficiaryVasp == "" {
			return nil, status.Error(codes.InvalidArgument, "if external demo is true, must specify beneficiary vasp")
		}

		wallet = &db.Wallet{
			Address: req.Beneficiary,
			Provider: db.VASP{
				Name: req.BeneficiaryVasp,
			},
			Vasp: s.vasp,
		}
	} else {
		// Lookup beneficiary wallet from the database and confirm it belongs to a remote RVASP
		wallet = &db.Wallet{}
		if err = s.db.LookupBeneficiary(req.Beneficiary).First(wallet).Error; err != nil {
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

// createTransaction returns a new Transaction from the originator account and
// beneficiary wallet, ready to be modified and/or stored in the database.
func (s *Server) createTransaction(originator db.Account, beneficiary db.Wallet) (*db.Transaction, error) {
	var originatorIdentity, beneficiaryIdentity db.Identity

	// Fetch originator identity record
	if err := s.db.LookupIdentity(originator.WalletAddress).FirstOrInit(&originatorIdentity, db.Identity{}).Error; err != nil {
		log.Error().Err(err).Msg("could not lookup originator identity")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup originator identity: %s", err)
	}

	// If originator identity does not exist then create it
	if originatorIdentity.ID == 0 {
		originatorIdentity.WalletAddress = originator.WalletAddress
		originatorIdentity.Vasp = s.vasp

		if err := s.db.Create(&originatorIdentity).Error; err != nil {
			log.Error().Err(err).Msg("could not save originator identity")
			return nil, status.Errorf(codes.FailedPrecondition, "could not save originator identity: %s", err)
		}
	}

	// Fetch beneficiary identity record
	if err := s.db.LookupIdentity(beneficiary.Address).FirstOrInit(&beneficiaryIdentity, db.Identity{}).Error; err != nil {
		log.Error().Err(err).Msg("could not lookup beneficiary identity")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup beneficiary identity: %s", err)
	}

	// If the beneficiary identity does not exist then create it
	if beneficiaryIdentity.ID == 0 {
		beneficiaryIdentity.WalletAddress = beneficiary.Address
		beneficiaryIdentity.Vasp = s.vasp

		if err := s.db.Create(&beneficiaryIdentity).Error; err != nil {
			log.Error().Err(err).Msg("could not save beneficiary identity")
			return nil, status.Errorf(codes.FailedPrecondition, "could not save beneficiary identity: %s", err)
		}
	}

	return &db.Transaction{
		Envelope:    uuid.New().String(),
		Originator:  originatorIdentity,
		Beneficiary: beneficiaryIdentity,
		State:       db.TransactionPending,
		Timestamp:   time.Now(),
		Vasp:        s.vasp,
	}, nil
}

// In the Basic Synchronous scenario:
// 1. The originator sends a request containing: the originator identity, the
//    beneficiary identity, and the complete transaction details.
// 2. The originator receives a response containing the complete payload and the
//    received_at timestamp.
// 3. The originator validates the payload and returns an error if necessary.
func (s *Server) basicSyncTransfer(req *pb.TransferRequest, account db.Account) (rep *pb.TransferReply, err error) {
	rep = &pb.TransferReply{}

	// Fetch the beneficiary Wallet
	var beneficiary *db.Wallet
	if beneficiary, err = s.fetchBeneficiaryWallet(req); err != nil {
		return nil, err
	}

	// Conduct a TRISADS lookup if necessary to get the endpoint
	var peer *peers.Peer
	if peer, err = s.peers.Search(beneficiary.Provider.Name); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return nil, status.Errorf(codes.Internal, "could not search peer from directory service: %s", err)
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return nil, status.Errorf(codes.FailedPrecondition, "could not exchange keys with remote peer: %s", err)
	}

	// Create a new Transaction in the database
	var xfer *db.Transaction
	if xfer, err = s.createTransaction(account, *beneficiary); err != nil {
		return nil, err
	}
	xfer.Account = account
	xfer.Amount = decimal.NewFromFloat32(req.Amount)
	xfer.Debit = true

	if err = s.db.Create(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save pending transaction")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save pending transaction: %s", err)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending++
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save originator account: %s", err)
	}

	// Create an identity and transaction payload for TRISA exchange
	transaction := &generic.Transaction{
		Txid:        fmt.Sprintf("%d", xfer.ID),
		Originator:  account.WalletAddress,
		Beneficiary: beneficiary.Address,
		Amount:      float64(req.Amount),
		Network:     "TestNet",
		Timestamp:   xfer.Timestamp.Format(time.RFC3339),
	}
	var identity *ivms101.IdentityPayload
	if identity, err = s.createIdentityPayload(account, beneficiary.Address); err != nil {
		return nil, err
	}

	var payload *protocol.Payload
	if payload, err = createTransferPayload(identity, transaction); err != nil {
		log.Error().Err(err).Msg("could not create transfer payload")
		return nil, status.Errorf(codes.Internal, "could not create transfer payload: %s", err)
	}

	// Secure the envelope with the remote beneficiary's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(xfer.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return nil, status.Errorf(codes.FailedPrecondition, "could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.trisa.sign))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	if payload.ReceivedAt == "" {
		log.Error().Msg("TRISA protocol error: received_at timestamp missing")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: received_at timestamp missing")
	}

	if identity, err = parseIdentityPayload(payload); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error: could not parse identity payload")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse identity payload: %s", err)
	}

	// Validate that the beneficiary identity matches the original request
	if identity.BeneficiaryVasp == nil || identity.BeneficiaryVasp.BeneficiaryVasp == nil {
		log.Error().Msg("TRISA protocol error: missing beneficiary vasp identity")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary vasp identity")
	}

	beneficiaryInfo := identity.Beneficiary
	if beneficiaryInfo == nil {
		log.Error().Msg("TRISA protocol error: missing beneficiary identity")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary identity")
	}

	if len(beneficiaryInfo.BeneficiaryPersons) != 1 {
		log.Error().Int("persons", len(beneficiaryInfo.BeneficiaryPersons)).Msg("TRISA protocol error: unexpected number of beneficiary persons")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: unexpected number of beneficiary persons")
	}

	if len(beneficiaryInfo.AccountNumbers) != 1 {
		log.Error().Int("accounts", len(beneficiaryInfo.AccountNumbers)).Msg("TRISA protocol error: unexpected number of beneficiary accounts")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: unexpected number of beneficiary accounts")
	}

	// TODO: Validate that the actual address was returned
	if beneficiaryInfo.AccountNumbers[0] == "" {
		log.Error().Msg("TRISA protocol error: missing beneficiary address")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary address")
	}

	if transaction, err = parseTransactionPayload(payload); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error: could not parse transaction payload")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse transaction payload: %s", err)
	}

	// Update the completed transaction and save to disk
	xfer.State = db.TransactionCompleted
	xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)

	// Serialize the identity information as JSON data
	var data []byte
	if data, err = json.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal IVMS 101 identity")
		return nil, status.Errorf(codes.Internal, "could not marshal IVMS 101 identity: %s", err)
	}
	xfer.Identity = string(data)

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save completed transaction")
		return nil, status.Errorf(codes.Internal, "could not save completed transaction: %s", err)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending--
	account.Completed++
	account.Balance.Sub(xfer.Amount)
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return nil, status.Errorf(codes.Internal, "could not save originator account: %s", err)
	}

	// Return the transfer response
	rep.Transaction = xfer.Proto()
	return rep, nil
}

// In the Partial Synchronous scenario:
// 1. The originator sends a request containing the originator identity and the
//    complete transaction details.
// 2. The originator receives a response containing the complete payload, the
//    beneficiary identity, and the received_at timestamp.
// 3. The originator validates the payload and returns an error if necessary.
func (s *Server) partialSyncTransfer(req *pb.TransferRequest, account db.Account) (rep *pb.TransferReply, err error) {
	rep = &pb.TransferReply{}

	// Fetch the beneficiary Wallet
	var beneficiary *db.Wallet
	if beneficiary, err = s.fetchBeneficiaryWallet(req); err != nil {
		return nil, err
	}

	// Conduct a TRISADS lookup if necessary to get the endpoint
	var peer *peers.Peer
	if peer, err = s.peers.Search(beneficiary.Provider.Name); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return nil, status.Errorf(codes.Internal, "could not search peer from directory service: %s", err)
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return nil, status.Errorf(codes.FailedPrecondition, "could not exchange keys with remote peer: %s", err)
	}

	// Create a new Transaction in the database
	var xfer *db.Transaction
	if xfer, err = s.createTransaction(account, *beneficiary); err != nil {
		return nil, err
	}
	xfer.Account = account
	xfer.Amount = decimal.NewFromFloat32(req.Amount)
	xfer.Debit = true

	if err = s.db.Create(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save pending transaction")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save pending transaction: %s", err)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending++
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save originator account: %s", err)
	}

	// Create an identity and transaction payload for TRISA exchange
	transaction := &generic.Transaction{
		Txid:        fmt.Sprintf("%d", xfer.ID),
		Originator:  account.WalletAddress,
		Beneficiary: beneficiary.Address,
		Amount:      float64(req.Amount),
		Network:     "TestNet",
		Timestamp:   xfer.Timestamp.Format(time.RFC3339),
	}
	var identity *ivms101.IdentityPayload
	if identity, err = s.createIdentityPayload(account, ""); err != nil {
		return nil, err
	}

	var payload *protocol.Payload
	if payload, err = createTransferPayload(identity, transaction); err != nil {
		log.Error().Err(err).Msg("could not create transfer payload")
		return nil, status.Errorf(codes.Internal, "could not create transfer payload: %s", err)
	}

	// Secure the envelope with the remote beneficiary's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(xfer.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return nil, status.Errorf(codes.FailedPrecondition, "could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.trisa.sign))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	if payload.ReceivedAt == "" {
		log.Error().Msg("TRISA protocol error: received_at timestamp missing")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: received_at timestamp missing")
	}

	if identity, err = parseIdentityPayload(payload); err != nil {
		log.Error().Err(err).Msg("could not parse identity payload")
		return nil, status.Errorf(codes.FailedPrecondition, "could not parse identity payload: %s", err)
	}

	if transaction, err = parseTransactionPayload(payload); err != nil {
		log.Error().Err(err).Msg("could not parse transaction payload")
		return nil, status.Errorf(codes.FailedPrecondition, "could not parse transaction payload: %s", err)
	}

	// Update the completed transaction and save to disk
	xfer.State = db.TransactionCompleted
	xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)

	// Serialize the identity information as JSON data
	var data []byte
	if data, err = json.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal IVMS 101 identity")
		return nil, status.Errorf(codes.Internal, "could not marshal IVMS 101 identity: %s", err)
	}
	xfer.Identity = string(data)

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save completed transaction")
		return nil, status.Errorf(codes.Internal, "could not save completed transaction: %s", err)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending--
	account.Completed++
	account.Balance.Sub(xfer.Amount)
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return nil, status.Errorf(codes.Internal, "could not save originator account: %s", err)
	}

	// Return the transfer response
	rep.Transaction = xfer.Proto()
	return rep, nil
}

// AsyncTransfer performs the first part of an asynchronous transfer with a beneficiary
// 1. The originator sends a request containing at least the originator identity and a
//    partial transaction which does not include the transaction ID.
// 2. The originator receives a pending protocol message with NotBefore and NotAfter
//    timestamps and updates the pending transaction in the database.
// 3. The originator waits for a response with the beneficiary information and a
//    received_at timestamp.
// 4. The originator echoes the response if the response was received within the time
//    window, otherwise it responds to the beneficiary with a canceled error.
// 5. The originator sends a new request with the transaction ID filled in.
// 6. The originator validates the echoed response from the beneficiary and returns
//    an error if necessary.
func (s *Server) AsyncTransfer(req *pb.TransferRequest, account db.Account) (rep *pb.TransferReply, err error) {
	rep = &pb.TransferReply{}

	// Fetch the beneficiary Wallet
	var beneficiary *db.Wallet
	if beneficiary, err = s.fetchBeneficiaryWallet(req); err != nil {
		return nil, err
	}

	// Conduct a TRISADS lookup if necessary to get the endpoint
	var peer *peers.Peer
	if peer, err = s.peers.Search(beneficiary.Provider.Name); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return nil, status.Errorf(codes.Internal, "could not search peer from directory service: %s", err)
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return nil, status.Errorf(codes.FailedPrecondition, "could not exchange keys with remote peer: %s", err)
	}

	// Create a new Transaction in the database
	var xfer *db.Transaction
	if xfer, err = s.createTransaction(account, *beneficiary); err != nil {
		return nil, err
	}
	xfer.Account = account
	xfer.Amount = decimal.NewFromFloat32(req.Amount)
	xfer.Debit = true

	if err = s.db.Create(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save created transaction")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save created transaction: %s", err)
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending++
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save originator account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save originator account: %s", err)
	}

	// Create an identity and transaction payload for TRISA exchange
	transaction := &generic.Transaction{
		Txid:        fmt.Sprintf("%d", xfer.ID),
		Originator:  account.WalletAddress,
		Beneficiary: beneficiary.Address,
		Amount:      float64(req.Amount),
		Network:     "TestNet",
		Timestamp:   xfer.Timestamp.Format(time.RFC3339),
	}
	var identity *ivms101.IdentityPayload
	if identity, err = s.createIdentityPayload(account, beneficiary.Address); err != nil {
		return nil, err
	}

	var payload *protocol.Payload
	if payload, err = createTransferPayload(identity, transaction); err != nil {
		log.Error().Err(err).Msg("could not create transfer payload")
		return nil, status.Errorf(codes.Internal, "could not create transfer payload: %s", err)
	}

	// Secure the envelope with the remote beneficiary's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(xfer.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return nil, status.Errorf(codes.FailedPrecondition, "could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.trisa.sign))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: %s", err)
	}

	// Parse the pending response
	var pending *generic.Pending
	if pending, err = parsePendingMessage(payload); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error: could not parse pending message")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse pending message: %s", err)
	}

	// Update the Transaction in the database with the pending timestamps
	if xfer.NotBefore, err = time.Parse(time.RFC3339, pending.ReplyNotBefore); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error: could not parse ReplyNotBefore timestamp")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse ReplyNotBefore timestamp: %s", err)
	}

	if xfer.NotAfter, err = time.Parse(time.RFC3339, pending.ReplyNotAfter); err != nil {
		log.Error().Err(err).Msg("TRISA protocol error: could not parse ReplyNotAfter timestamp")
		return nil, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: could not parse ReplyNotAfter timestamp: %s", err)
	}
	xfer.State = db.TransactionPending

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save pending transaction")
		return nil, status.Errorf(codes.Internal, "could not save pending transaction: %s", err)
	}

	rep.Transaction = xfer.Proto()
	return rep, nil
}

// In the Rejected Asynchronous scenario:
// 1. The originator sends a request containing at least the originator identity and a
//    partial transaction which does not include the transaction ID.
// 2. The originator receives a pending protocol message with NotBefore and NotAfter
//    timestamps.
// 3. The originator waits for a protocol reject message and stops the transaction.
func (s *Server) rejectedAsyncTransfer(req *pb.TransferRequest, account db.Account) (rep *pb.TransferReply, err error) {
	// TODO: Implement this
	return nil, status.Error(codes.Unimplemented, "rejected_async transfer is not implemented")
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
		log.Error().Err(err).Msg("could not lookup account")
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
			pb.Errorf(pb.ErrInternal, "could not exchange keyrs with remote peer"),
		)
	}

	// Prepare the transaction
	// Save the pending transaction and increment the accounts pending field
	xfer := db.Transaction{
		Envelope: uuid.New().String(),
		Account:  account,
		Amount:   decimal.NewFromFloat32(transfer.Amount),
		Debit:    true,
		State:    db.TransactionPending,
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
	xfer.State = db.TransactionCompleted
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
	account.Balance.Sub(xfer.Amount)
	if err = s.db.Save(&account).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction on account")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	message = fmt.Sprintf("transaction %04d complete: %s transfered from %s to %s", xfer.ID, xfer.Amount.String(), xfer.Originator.WalletAddress, xfer.Beneficiary.WalletAddress)
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

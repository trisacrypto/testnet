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
	"github.com/trisacrypto/directory/pkg/gds/logger"
	"github.com/trisacrypto/testnet/pkg"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/handler"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/driver/sqlite"
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
func New(conf *Settings) (s *Server, err error) {
	if conf == nil {
		if conf, err = Config(); err != nil {
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
	if s.db, err = gorm.Open(sqlite.Open(conf.DatabaseDSN), &gorm.Config{}); err != nil {
		return nil, err
	}

	if err = MigrateDB(s.db); err != nil {
		return nil, err
	}

	// TODO: mark the VASP local based on name or configuration rather than erroring
	if err = s.db.Where("is_local = ?", true).First(&s.vasp).Error; err != nil {
		return nil, fmt.Errorf("could not fetch local VASP info from database: %s", err)
	}

	if s.conf.Name != s.vasp.Name {
		return nil, fmt.Errorf("expected name %q but have database name %q", s.conf.Name, s.vasp.Name)
	}

	// Create the TRISA service
	if s.trisa, err = NewTRISA(s); err != nil {
		return nil, fmt.Errorf("could not create TRISA service: %s", err)
	}

	// Create the remote peers using the same credentials as the TRISA service
	s.peers = peers.New(s.trisa.certs, s.trisa.chain, s.conf.DirectoryServiceURL)
	s.updates = NewUpdateManager()
	return s, nil
}

// Server implements the GRPC TRISAIntegration and TRISADemo services.
type Server struct {
	pb.UnimplementedTRISADemoServer
	pb.UnimplementedTRISAIntegrationServer
	conf    *Settings
	srv     *grpc.Server
	db      *gorm.DB
	vasp    VASP
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
func (s *Server) Transfer(ctx context.Context, req *pb.TransferRequest) (rep *pb.TransferReply, err error) {
	rep = &pb.TransferReply{}

	// Get originator account and confirm it belongs to this RVASP
	var account Account
	if err = LookupAccount(s.db, req.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info().Str("account", req.Account).Msg("not found")
			return nil, status.Error(codes.NotFound, "account not found")
		}
		log.Error().Err(err).Msg("could not lookup account")
		return nil, status.Errorf(codes.FailedPrecondition, "could not lookup account: %s", err)
	}

	// Identify the beneficiary either using the demo database or the directory service
	var beneficiary Wallet
	if req.ExternalDemo {
		if req.BeneficiaryVasp == "" {
			return nil, status.Error(codes.InvalidArgument, "if external demo is true, must specify beneficiary vasp")
		}

		beneficiary = Wallet{
			Address: req.Beneficiary,
			Provider: VASP{
				Name: req.BeneficiaryVasp,
			},
		}
	} else {
		// Lookup beneficiary wallet and confirm it belongs to a remote RVASP
		if err = LookupBeneficiary(s.db, req.Beneficiary).First(&beneficiary).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Info().Str("beneficiary", req.Beneficiary).Msg("not found")
				return nil, status.Error(codes.NotFound, "beneficiary not found (use external_demo?)")
			}
			log.Error().Err(err).Msg("could not lookup beneficiary")
			return nil, status.Errorf(codes.FailedPrecondition, "could not lookup beneficiary: %s", err)
		}

		if req.CheckBeneficiary {
			if req.BeneficiaryVasp != beneficiary.Provider.Name {
				log.Warn().
					Str("expected", req.BeneficiaryVasp).
					Str("actual", beneficiary.Provider.Name).
					Msg("check beneficiary failed")
				return nil, status.Error(codes.InvalidArgument, "beneficiary wallet does not match beneficiary VASP")
			}

		}
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

	// Save the pending transaction and increment the accounts pending field
	xfer := Transaction{
		Envelope:  uuid.New().String(),
		Account:   account,
		Amount:    decimal.NewFromFloat32(req.Amount),
		Debit:     true,
		Completed: false,
		Timestamp: time.Now(),
	}

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction")
		return nil, status.Errorf(codes.FailedPrecondition, "could not save transaction: %s", err)
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
	identity := &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
	}
	if identity.OriginatingVasp.OriginatingVasp, err = s.vasp.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load originator vasp")
		return nil, status.Errorf(codes.Internal, "could not load originator vasp: %s", err)
	}

	identity.Originator = &ivms101.Originator{
		OriginatorPersons: make([]*ivms101.Person, 0, 1),
		AccountNumbers:    []string{account.WalletAddress},
	}
	var originator *ivms101.Person
	if originator, err = account.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load originator identity")
		return nil, status.Errorf(codes.Internal, "could not load originator identity: %s", err)
	}
	identity.Originator.OriginatorPersons = append(identity.Originator.OriginatorPersons, originator)

	payload := &protocol.Payload{}
	if payload.Transaction, err = anypb.New(transaction); err != nil {
		log.Error().Err(err).Msg("could not dump payload transaction")
		return nil, status.Errorf(codes.Internal, "could not dump payload transaction: %s", err)
	}
	if payload.Identity, err = anypb.New(identity); err != nil {
		log.Error().Err(err).Msg("could not dump payload identity")
		return nil, status.Errorf(codes.Internal, "could not dump payload identity: %s", err)
	}

	// Secure the envelope with the remote beneficiary's signing keys
	var envelope *protocol.SecureEnvelope
	if envelope, err = handler.New(xfer.Envelope, payload, nil).Seal(signKey); err != nil {
		log.Error().Err(err).Msg("could not create or sign secure envelope")
		return nil, status.Errorf(codes.FailedPrecondition, "could not create or sign secure envelope: %s", err)
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if envelope, err = peer.Transfer(envelope); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return nil, status.Errorf(codes.FailedPrecondition, "could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	var opened *handler.Envelope
	if opened, err = handler.Open(envelope, s.trisa.sign); err != nil {
		log.Error().Err(err).Msg("could not unseal TRISA response")
		return nil, status.Errorf(codes.FailedPrecondition, "could not unseal TRISA response: %s", err)
	}

	// Verify the contents of the response
	payload = opened.Payload
	if payload.Identity == nil || payload.Transaction == nil {
		// Check if we've received a confirmation receipt
		if payload.Transaction != nil {
			switch payload.Transaction.TypeUrl {
			case "type.googleapis.com/trisa.data.generic.v1beta1.ConfirmationReceipt":
				receipt := &generic.ConfirmationReceipt{}
				if err = payload.Transaction.UnmarshalTo(receipt); err != nil {
					log.Error().Err(err).Msg("could not unmarshal confirmation receipt")
					return nil, status.Errorf(codes.FailedPrecondition, "could not unmarshal confirmation receipt: %s", err)
				}
				log.Info().
					Str("envelope", receipt.EnvelopeId).
					Str("received_by", receipt.ReceivedBy).
					Str("received_at", receipt.ReceivedAt).
					Str("message", receipt.Message).
					Msg("confirmation receipt received")
				err = fmt.Errorf("received confirmation ID %s from %s: %s", receipt.EnvelopeId, receipt.ReceivedBy, receipt.Message)
				return nil, status.Error(codes.Unimplemented, err.Error())
			case "type.googleapis.com/ciphertrace.apis.traveler.common.v1.ConfirmationReceipt":
				log.Info().Msg("received Traveler confirmation receipt")
				return nil, status.Error(codes.Unimplemented, "received confirmation receipt, transaction could not be completed synchronously")
			default:
				log.Error().Str("type", payload.Transaction.TypeUrl).Msg("unknown confirmation receipt type")
				return nil, status.Error(codes.FailedPrecondition, "could not parse confirmation receipt")
			}
		}

		log.Warn().Msg("did not receive identity or transaction")
		return nil, status.Error(codes.FailedPrecondition, "no identity or transaction returned")
	}

	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unsupported identity type")
		return nil, status.Errorf(codes.FailedPrecondition, "unsupported identity type for rVASP: %q", payload.Identity.TypeUrl)
	}

	if payload.Transaction.TypeUrl != "type.googleapis.com/trisa.data.generic.v1beta1.Transaction" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported transaction type")
		return nil, status.Errorf(codes.FailedPrecondition, "unsupported identity type for rVASP: %q", payload.Transaction.TypeUrl)
	}

	identity = &ivms101.IdentityPayload{}
	transaction = &generic.Transaction{}
	if err = payload.Identity.UnmarshalTo(identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity")
		return nil, status.Errorf(codes.FailedPrecondition, "could not unmarshal identity: %s", err)
	}
	if err = payload.Transaction.UnmarshalTo(transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal transaction")
		return nil, status.Errorf(codes.FailedPrecondition, "could not unmarshal transaction: %s", err)
	}

	// Update the completed transaction and save to disk
	xfer.Beneficiary = Identity{
		WalletAddress: transaction.Beneficiary,
	}
	xfer.Completed = true
	xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)

	// Serialize the identity information as JSON data
	var data []byte
	if data, err = json.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not marshal IVMS 101 identity")
		return nil, status.Errorf(codes.Internal, "could not marshal IVMS 101 identity: %s", err)
	}
	xfer.Identity = string(data)

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction")
		return nil, status.Errorf(codes.Internal, "could not save transaction: %s", err)
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

// AccountStatus is a demo RPC to allow demo clients to fetch their recent transactions.
func (s *Server) AccountStatus(ctx context.Context, req *pb.AccountRequest) (rep *pb.AccountReply, err error) {
	rep = &pb.AccountReply{}

	// Lookup the account in the database
	var account Account
	if err = LookupAccount(s.db, req.Account).First(&account).Error; err != nil {
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
		var transactions []Transaction
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
	msg := fmt.Sprintf("starting transaction of %0.2f from %s to %s", transfer.Amount, transfer.Account, transfer.Beneficiary)
	s.updates.Broadcast(req.Id, msg, pb.MessageCategory_LEDGER)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Handle Demo UI errors before the account lookup
	if transfer.OriginatingVasp != "" && transfer.OriginatingVasp != s.vasp.Name {
		log.Info().Str("requested", transfer.OriginatingVasp).Str("local", s.vasp.Name).Msg("requested originator does not match local VASP")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrWrongVASP, "message sent to the wrong originator VASP"),
		)
	}

	// Lookup the account associated with the transfer originator
	var account Account
	if err = LookupAccount(s.db, transfer.Account).First(&account).Error; err != nil {
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
	var beneficiary Wallet
	if err = LookupBeneficiary(s.db, transfer.Beneficiary).First(&beneficiary).Error; err != nil {
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
	xfer := Transaction{
		Envelope:  uuid.New().String(),
		Account:   account,
		Amount:    decimal.NewFromFloat32(transfer.Amount),
		Debit:     true,
		Completed: false,
	}

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not save transaction"),
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
	var envelope *protocol.SecureEnvelope
	if envelope, err = handler.New(xfer.Envelope, payload, nil).Seal(signKey); err != nil {
		log.Error().Err(err).Msg("could not create or sign secure envelope")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not create or sign secure envelope"),
		)
	}

	s.updates.Broadcast(req.Id, fmt.Sprintf("secure envelope %s sealed: encrypted with AES-GCM and RSA - sending ...", envelope.Id), pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Conduct the TRISA transaction, handle errors and send back to user
	if envelope, err = peer.Transfer(envelope); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	s.updates.Broadcast(req.Id, fmt.Sprintf("received %s information exchange reply from %s", envelope.Id, peer.String()), pb.MessageCategory_TRISAP2P)
	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)

	// Open the response envelope with local private keys
	var opened *handler.Envelope
	if opened, err = handler.Open(envelope, s.trisa.sign); err != nil {
		log.Error().Err(err).Msg("could not unseal TRISA response")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	// Verify the contents of the response
	payload = opened.Payload
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
	xfer.Beneficiary = Identity{
		WalletAddress: transaction.Beneficiary,
	}
	xfer.Completed = true
	xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)

	// Serialize the identity information as JSON data
	var data []byte
	if data, err = json.Marshal(identity); err != nil {
		log.Error().Err(err).Msg("could not save transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, "could not marshal IVMS 101 identity"),
		)
	}
	xfer.Identity = string(data)

	if err = s.db.Save(&xfer).Error; err != nil {
		log.Error().Err(err).Msg("could not save transaction")
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
		log.Error().Err(err).Msg("could not save transaction")
		return s.updates.SendTransferError(client, req.Id,
			pb.Errorf(pb.ErrInternal, err.Error()),
		)
	}

	msg = fmt.Sprintf("transaction %04d complete: %s transfered from %s to %s", xfer.ID, xfer.Amount.String(), xfer.Originator.WalletAddress, xfer.Beneficiary.WalletAddress)
	s.updates.Broadcast(req.Id, msg, pb.MessageCategory_BLOCKCHAIN)
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

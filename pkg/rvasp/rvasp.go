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

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/trisacrypto/testnet/pkg"
	"github.com/trisacrypto/testnet/pkg/ivms101"
	api "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/trisacrypto/testnet/pkg/trisa/handler"
	"github.com/trisacrypto/testnet/pkg/trisa/peers"
	protocol "github.com/trisacrypto/testnet/pkg/trisa/protocol/v1alpha1"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
	return s, nil
}

// Server implements the GRPC TRISAIntegration and TRISADemo services.
type Server struct {
	pb.UnimplementedTRISADemoServer
	pb.UnimplementedTRISAIntegrationServer
	conf  *Settings
	srv   *grpc.Server
	db    *gorm.DB
	vasp  VASP
	trisa *TRISA
	echan chan error
	peers *peers.Peers
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
			rep.Error = pb.Errorf(pb.ErrNotFound, "account not found")
			log.Info().Str("account", req.Account).Msg("not found")
			return rep, nil
		}
		log.Error().Err(err).Msg("could not lookup account")
		return nil, err
	}

	// Lookup beneficiary wallet and confirm it belongs to a remote RVASP
	var beneficiary Wallet
	if err = LookupBeneficiary(s.db, req.Beneficiary).First(&beneficiary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rep.Error = pb.Errorf(pb.ErrNotFound, "beneficiary not found")
			log.Info().Str("beneficiary", req.Beneficiary).Msg("not found")
			return rep, nil
		}
		log.Error().Err(err).Msg("could not lookup beneficiary")
		return nil, err
	}

	if req.CheckBeneficiary {
		if req.BeneficiaryVasp != beneficiary.Provider.Name {
			rep.Error = pb.Errorf(pb.ErrWrongVASP, "beneficiary wallet does not match beneficiary VASP")
			log.Info().
				Str("expected", req.BeneficiaryVasp).
				Str("actual", beneficiary.Provider.Name).
				Msg("check beneficiary failed")
			return rep, nil
		}

	}

	// Conduct a TRISADS lookup if necessary to get the endpoint
	var peer *peers.Peer
	if peer, err = s.peers.Search(beneficiary.Provider.Name); err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		log.Error().Err(err).Msg("could not search peer from directory service")
		return rep, nil
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(false); err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return rep, nil
	}

	// Save the pending transaction and increment the accounts pending field
	xfer := Transaction{
		Envelope:  uuid.New().String(),
		Account:   account,
		Amount:    decimal.NewFromFloat32(req.Amount),
		Debit:     true,
		Completed: false,
	}

	if err = s.db.Save(&xfer).Error; err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		log.Error().Err(err).Msg("could not save transaction")
		return rep, nil
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending++
	if err = s.db.Save(&account).Error; err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		log.Error().Err(err).Msg("could not save originator account")
		return rep, nil
	}

	// Create an identity and transaction payload for TRISA exchange
	transaction := &api.Transaction{
		Originator: &api.Account{
			WalletAddress: account.WalletAddress,
			Email:         account.Email,
			Provider:      s.vasp.Name,
		},
		Beneficiary: &api.Account{
			WalletAddress: beneficiary.Address,
		},
		Amount: req.Amount,
	}
	identity := &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
	}
	if identity.OriginatingVasp.OriginatingVasp, err = s.vasp.LoadIdentity(); err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, "could not load originator vasp")
		log.Error().Err(err).Msg("could not load originator vasp")
		return rep, nil
	}

	identity.Originator = &ivms101.Originator{
		OriginatorPersons: make([]*ivms101.Person, 0, 1),
		AccountNumbers:    []string{account.WalletAddress},
	}
	var originator *ivms101.Person
	if originator, err = account.LoadIdentity(); err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, "could not load originator identity")
		log.Error().Err(err).Msg("could not load originator identity")
		return rep, nil
	}
	identity.Originator.OriginatorPersons = append(identity.Originator.OriginatorPersons, originator)

	payload := &protocol.Payload{}
	if payload.Transaction, err = ptypes.MarshalAny(transaction); err != nil {
		log.Error().Err(err).Msg("could not dump payload transaction")
		rep.Error = pb.Errorf(pb.ErrInternal, "could not dump payload transaction")
		return rep, nil
	}
	if payload.Identity, err = ptypes.MarshalAny(identity); err != nil {
		log.Error().Err(err).Msg("could not dump payload identity")
		rep.Error = pb.Errorf(pb.ErrInternal, "could not dump payload identity")
		return rep, nil
	}

	// Secure the envelope with the remote beneficiary's signing keys
	var envelope *protocol.SecureEnvelope
	if envelope, err = handler.New(xfer.Envelope, payload, nil).Seal(signKey); err != nil {
		log.Error().Err(err).Msg("could not create or sign secure envelope")
		rep.Error = pb.Errorf(pb.ErrInternal, "could not create or sign secure envelope")
		return rep, nil
	}

	// Conduct the TRISA transaction, handle errors and send back to user
	if envelope, err = peer.Transfer(envelope); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		return rep, nil
	}

	// Open the response envelope with local private keys
	var opened *handler.Envelope
	if opened, err = handler.Open(envelope, s.trisa.sign); err != nil {
		log.Error().Err(err).Msg("could not unseal TRISA response")
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		return rep, nil
	}

	// Verify the contents of the response
	payload = opened.Payload
	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unsupported identity type")
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		return rep, nil
	}

	if payload.Transaction.TypeUrl != "type.googleapis.com/rvasp.v1.Transaction" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported transaction type")
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		return rep, nil
	}

	identity = &ivms101.IdentityPayload{}
	transaction = &api.Transaction{}
	if err = ptypes.UnmarshalAny(payload.Identity, identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity")
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		return rep, nil
	}
	if err = ptypes.UnmarshalAny(payload.Transaction, transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal transaction")
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		return rep, nil
	}

	// Update the completed transaction and save to disk
	xfer.Beneficiary = Identity{
		WalletAddress: transaction.Beneficiary.WalletAddress,
		Email:         transaction.Beneficiary.Email,
		Provider:      transaction.Beneficiary.Provider,
	}
	xfer.Completed = true
	xfer.Timestamp, _ = time.Parse(time.RFC3339, transaction.Timestamp)

	// Serialize the identity information as JSON data
	var data []byte
	if data, err = json.Marshal(identity); err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, "could not marshal IVMS 101 identity")
		log.Error().Err(err).Msg("could not save transaction")
		return rep, nil
	}
	xfer.Identity = string(data)

	if err = s.db.Save(&xfer).Error; err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		log.Error().Err(err).Msg("could not save transaction")
		return rep, nil
	}

	// Save the pending transaction on the account
	// TODO: remove pending transactions
	account.Pending--
	account.Completed++
	account.Balance.Sub(xfer.Amount)
	if err = s.db.Save(&account).Error; err != nil {
		rep.Error = pb.Errorf(pb.ErrInternal, err.Error())
		log.Error().Err(err).Msg("could not save originator account")
		return rep, nil
	}

	// Return the transfer response
	rep.Transaction = transaction
	rep.Transaction.Envelope = xfer.Envelope
	rep.Transaction.Identity = xfer.Identity
	return rep, nil
}

// AccountStatus is a demo RPC to allow demo clients to fetch their recent transactions.
func (s *Server) AccountStatus(ctx context.Context, req *pb.AccountRequest) (rep *pb.AccountReply, err error) {
	rep = &pb.AccountReply{}

	// Lookup the account in the database
	var account Account
	if err = LookupAccount(s.db, req.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rep.Error = pb.Errorf(pb.ErrNotFound, "account not found")
			log.Info().Err(err).Msg("account not found")
			return rep, nil
		}
		log.Error().Err(err).Msg("could not lookup account")
		return nil, err
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
			return nil, err
		}

		rep.Transactions = make([]*pb.Transaction, 0, len(transactions))
		for _, transaction := range transactions {
			rep.Transactions = append(rep.Transactions, &pb.Transaction{
				Originator: &pb.Account{
					WalletAddress: transaction.Originator.WalletAddress,
					Email:         transaction.Originator.Email,
					Provider:      transaction.Originator.Provider,
				},
				Beneficiary: &pb.Account{
					WalletAddress: transaction.Beneficiary.WalletAddress,
					Email:         transaction.Beneficiary.Email,
					Provider:      transaction.Beneficiary.Provider,
				},
				Amount:    transaction.AmountFloat(),
				Timestamp: transaction.Timestamp.Format(time.RFC3339),
				Envelope:  transaction.Envelope,
				Identity:  transaction.Identity,
			})
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
			log.Info().Str("client", client).Msg("connected to live updates")
		} else if client != req.Client {
			log.Warn().Str("request from", req.Client).Str("client stream", client).Msg("unexpected client")
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
			if err = stream.Send(ack); err != nil {
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

			if err = stream.Send(ack); err != nil {
				log.Error().Err(err).Str("client", client).Msg("could not send message")
				return err
			}
		case pb.RPC_TRANSFER:
			// HACK: simulate the TRISA process as a quick stub to unblock the front end
			if err = s.simulateTRISA(stream, req, client); err != nil {
				log.Error().Err(err).Msg("could not simulate TRISA")
				return err
			}
		}
	}
}

func (s *Server) simulateTRISA(stream pb.TRISADemo_LiveUpdatesServer, req *pb.Command, client string) (err error) {
	// Create stream updater context for sending live updates back to client
	updater := newStreamUpdater(stream, req, client)

	// Get the transfer from the original command, will panic if nil
	transfer := req.GetTransfer()

	// Handle Demo UI errors before the account lookup
	if transfer.OriginatingVasp != s.vasp.Name {
		rep := &pb.Message{
			Type:      pb.RPC_TRANSFER,
			Id:        req.Id,
			Timestamp: time.Now().Format(time.RFC3339),
			Category:  pb.MessageCategory_ERROR,
			Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
				Error: pb.Errorf(pb.ErrWrongVASP, "message sent to the wrong originator VASP"),
			}},
		}
		if err = stream.Send(rep); err != nil {
			return fmt.Errorf("could not send transfer reply to %q: %s", client, err)
		}
		return nil
	}

	// Lookup the account associated with the transfer originator
	var account Account
	if err = LookupAccount(s.db, transfer.Account).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rep := &pb.Message{
				Type:      pb.RPC_TRANSFER,
				Id:        req.Id,
				Timestamp: time.Now().Format(time.RFC3339),
				Category:  pb.MessageCategory_ERROR,
				Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
					Error: pb.Errorf(pb.ErrNotFound, "account not found"),
				}},
			}

			if err = stream.Send(rep); err != nil {
				return fmt.Errorf("could not send transfer reply to %q: %s", client, err)
			}
			return nil
		}
		return fmt.Errorf("could not fetch account: %s", err)
	}
	if err = updater.send(fmt.Sprintf("account %d accessed successfully", account.ID), pb.MessageCategory_LEDGER); err != nil {
		return err
	}

	// Lookup the wallet of the beneficiary
	var beneficiary Wallet
	if err = LookupBeneficiary(s.db, transfer.Beneficiary).First(&beneficiary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rep := &pb.Message{
				Type:      pb.RPC_TRANSFER,
				Id:        req.Id,
				Timestamp: time.Now().Format(time.RFC3339),
				Category:  pb.MessageCategory_ERROR,
				Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
					Error: pb.Errorf(pb.ErrNotFound, "beneficiary wallet not found"),
				}},
			}

			if err = stream.Send(rep); err != nil {
				return fmt.Errorf("could not send transfer reply to %q: %s", client, err)
			}
			return nil
		}
		return fmt.Errorf("could not fetch beneficiary wallet: %s", err)
	}

	if transfer.CheckBeneficiary {
		if transfer.BeneficiaryVasp != beneficiary.Provider.Name {
			rep := &pb.Message{
				Type:      pb.RPC_TRANSFER,
				Id:        req.Id,
				Timestamp: time.Now().Format(time.RFC3339),
				Category:  pb.MessageCategory_ERROR,
				Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
					Error: pb.Errorf(pb.ErrWrongVASP, "beneficiary wallet does not match beneficiary vasp"),
				}},
			}
			if err = stream.Send(rep); err != nil {
				return fmt.Errorf("could not send transfer reply to %q: %s", client, err)
			}
			return nil
		}
	}

	if err = updater.send(fmt.Sprintf("wallet %s (%s) provided by %s", beneficiary.Address, beneficiary.Email, beneficiary.Provider.Name), pb.MessageCategory_BLOCKCHAIN); err != nil {
		return err
	}

	if err = updater.send("beginning TRISA protocol for identity exchange", pb.MessageCategory_TRISAP2P); err != nil {
		return err
	}

	if err = updater.send("VASP public key not cached, looking up TRISA directory service", pb.MessageCategory_TRISADS); err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Int63n(1800)) * time.Millisecond)
	if err = updater.send("sending handshake request to [endpoint]", pb.MessageCategory_TRISAP2P); err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Int63n(2200)) * time.Millisecond)
	if err = updater.send("[vasp] verified, secure TRISA connection established", pb.MessageCategory_TRISAP2P); err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Int63n(1800)) * time.Millisecond)
	if err = updater.send(fmt.Sprintf("identity for beneficiary %q confirmed - beginning transaction", beneficiary.Email), pb.MessageCategory_BLOCKCHAIN); err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Int63n(6200)) * time.Millisecond)
	if err = updater.send("transaction appended to blockchain, sending hash to [endpoint]", pb.MessageCategory_BLOCKCHAIN); err != nil {
		return err
	}

	time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	rep := &pb.Message{
		Type:      pb.RPC_TRANSFER,
		Id:        req.Id,
		Timestamp: time.Now().Format(time.RFC3339),
		Category:  pb.MessageCategory_LEDGER,
		Reply: &pb.Message_Transfer{Transfer: &pb.TransferReply{
			Transaction: &pb.Transaction{
				Originator: &pb.Account{
					WalletAddress: account.WalletAddress,
					Email:         account.Email,
					Provider:      s.vasp.IVMS101,
				},
				Beneficiary: &pb.Account{
					WalletAddress: beneficiary.Address,
					Email:         beneficiary.Email,
					Provider:      "[simulated]",
				},
				Amount:    transfer.Amount,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}},
	}

	if err = stream.Send(rep); err != nil {
		return fmt.Errorf("could not send transfer reply to %q: %s", client, err)
	}

	return nil
}

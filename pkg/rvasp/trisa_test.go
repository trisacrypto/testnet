package rvasp_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"github.com/trisacrypto/testnet/pkg/rvasp"
	"github.com/trisacrypto/testnet/pkg/rvasp/bufconn"
	"github.com/trisacrypto/testnet/pkg/rvasp/config"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/envelope"
	"github.com/trisacrypto/trisa/pkg/trisa/mtls"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	bufSize      = 1024 * 1024
	fixturesPath = "fixtures"
)

// rVASPTestSuite tests interactions with the rvasp servers to ensure that they return
// the expected error codes and messages.
type rVASPTestSuite struct {
	suite.Suite
	grpc     *bufconn.GRPCListener
	db       sqlmock.Sqlmock
	trisa    *rvasp.TRISA
	certs    *trust.Provider
	chain    trust.ProviderPool
	peers    *peers.Peers
	conf     *config.Config
	vasps    []db.VASP
	wallets  []db.Wallet
	accounts []db.Account
}

func TestRVASP(t *testing.T) {
	suite.Run(t, new(rVASPTestSuite))
}

func (s *rVASPTestSuite) SetupSuite() {
	var err error
	require := s.Require()

	s.conf, err = config.New()
	require.NoError(err)
	certPath := filepath.Join("testdata", "cert.pem")
	s.conf.CertPath = certPath
	s.conf.TrustChainPath = certPath

	s.vasps, err = db.LoadVASPs(fixturesPath)
	require.NoError(err, "could not load VASP fixtures")
	require.Greater(len(s.vasps), 0, "no VASPs loaded")

	s.wallets, s.accounts, err = db.LoadWallets(fixturesPath)
	require.NoError(err, "could not load wallet fixtures")
	require.Greater(len(s.wallets), 0, "no wallets loaded")
	require.Greater(len(s.accounts), 0, "no accounts loaded")
}

func (s *rVASPTestSuite) BeforeTest(suiteName, testName string) {
	var err error
	require := s.Require()
	s.trisa, s.peers, s.db, s.certs, s.chain, err = rvasp.NewTRISAMock(s.conf)
	require.NoError(err)

	s.grpc = bufconn.New(bufSize)
	go s.trisa.Run(s.grpc.Listener)
}

func (s *rVASPTestSuite) AfterTest(suiteName, testName string) {
	s.db.ExpectationsWereMet()
	s.trisa.Shutdown()
	if s.grpc != nil {
		s.grpc.Release()
	}
}

// expectStandardQuery preloads a query that returns a single row with an id as a response.
func expectStandardQuery(db sqlmock.Sqlmock, kind string) {
	db.ExpectQuery(kind).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
}

// createIdentityPayload which has a valid beneficiary identity and can be used for
// SyncRequire transfers.
func (s *rVASPTestSuite) createIdentityPayload() (identity *ivms101.IdentityPayload) {
	require := s.Require()

	identity = &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
		Beneficiary:     &ivms101.Beneficiary{},
		BeneficiaryVasp: &ivms101.BeneficiaryVasp{},
	}

	// For unit testing it does not matter which VASP fixture is used although it must
	// have a valid ivms101.LegalPerson
	var err error
	identity.BeneficiaryVasp.BeneficiaryVasp, err = s.vasps[0].LoadIdentity()
	require.NoError(err, "could not load beneficiary VASP identity")
	require.NoError(identity.BeneficiaryVasp.BeneficiaryVasp.GetLegalPerson().Validate(), "VASP ivms101.LegalPerson fixture failed validation")

	// The account fixture must have a valid ivms101.NaturalPerson
	beneficiary, err := s.accounts[0].LoadIdentity()
	require.NoError(err, "could not load beneficiary account identity")
	require.NoError(beneficiary.GetNaturalPerson().Validate(), "account ivms101.NaturalPerson fixture failed validation")

	identity.Beneficiary.BeneficiaryPersons = []*ivms101.Person{beneficiary}
	identity.Beneficiary.AccountNumbers = []string{s.accounts[0].WalletAddress}

	return identity
}

// Test that the TRISA server returns a valid envelope when a valid request is sent for
// SyncRequire.
func (s *rVASPTestSuite) TestValidTransfer() {
	var err error
	require := s.Require()

	originatorAddress := "alice@alicevasp.us"
	beneficiaryAddress := "george@bobvasp.co.uk"

	// Create the request envelope
	payload := &protocol.Payload{
		SentAt: time.Now().Format(time.RFC3339),
	}

	identity := s.createIdentityPayload()
	payload.Identity, err = anypb.New(identity)
	require.NoError(err)

	transaction := &generic.Transaction{
		Originator:  originatorAddress,
		Beneficiary: beneficiaryAddress,
	}
	payload.Transaction, err = anypb.New(transaction)
	require.NoError(err)

	// Preload the beneficiary address and policy fetches
	s.db.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"wallet_address"}).AddRow(beneficiaryAddress))
	s.db.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"beneficiary_policy"}).AddRow("SyncRequire"))

	// Preload the transaction lookup
	expectStandardQuery(s.db, "SELECT")

	// Preload the beneficiary account insert
	s.db.ExpectBegin()
	expectStandardQuery(s.db, "INSERT")
	s.db.ExpectCommit()

	// Account record update
	s.db.ExpectBegin()
	s.db.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	s.db.ExpectCommit()

	// Transaction record update
	s.db.ExpectBegin()
	s.db.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	s.db.ExpectCommit()

	// Seal the envelope using the public key
	key, err := s.certs.GetRSAKeys()
	require.NoError(err)
	msg, reject, err := envelope.Seal(payload, envelope.WithRSAPublicKey(&key.PublicKey))
	require.NoError(err)
	require.Nil(reject)

	// Start the gRPC client
	creds, err := mtls.ClientCreds("localhost", s.certs, s.chain)
	require.NoError(err)
	require.NoError(s.grpc.Connect(creds))
	defer s.grpc.Close()
	client := protocol.NewTRISANetworkClient(s.grpc.Conn)

	// Do the request
	response, err := client.Transfer(context.Background(), msg)
	require.NoError(err)
	require.NotNil(response)

	// Verify that a valid envelope was returned
	reject, isErr := envelope.Check(response)
	require.Nil(reject)
	require.False(isErr)

	// Decrypt the response envelope
	payload, reject, err = envelope.Open(response, envelope.WithRSAPrivateKey(key))
	require.NoError(err)
	require.Nil(reject)
	require.NotNil(payload)

	// Check for a valid response
	require.NotNil(payload.Identity)
	require.NotNil(payload.Transaction)
	require.NotEmpty(payload.SentAt)
	require.NotEmpty(payload.ReceivedAt)

	// Validate the transaction payload
	protoMsg, err := payload.Transaction.UnmarshalNew()
	require.NoError(err)
	actual, ok := protoMsg.(*generic.Transaction)
	require.True(ok)
	require.Equal(originatorAddress, actual.Originator)
	require.Equal(beneficiaryAddress, actual.Beneficiary)
}

// Test that the TRISA server sends back an error when an invalid request is sent for
// SyncRequire.
func (s *rVASPTestSuite) TestInvalidTransfer() {
	var err error
	require := s.Require()

	originatorAddress := "alice@alicevasp.us"
	beneficiaryAddress := "george@bobvasp.co.uk"

	// Create the request envelope with missing beneficiary info
	payload := &protocol.Payload{
		SentAt: time.Now().Format(time.RFC3339),
	}

	identity := &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
		Beneficiary:     &ivms101.Beneficiary{},
		BeneficiaryVasp: &ivms101.BeneficiaryVasp{},
	}
	payload.Identity, err = anypb.New(identity)
	require.NoError(err)

	transaction := &generic.Transaction{
		Originator:  originatorAddress,
		Beneficiary: beneficiaryAddress,
	}
	payload.Transaction, err = anypb.New(transaction)
	require.NoError(err)

	// Preload the beneficiary address and policy fetches
	s.db.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"wallet_address"}).AddRow(beneficiaryAddress))
	s.db.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"beneficiary_policy"}).AddRow("SyncRequire"))

	// Preload the transaction lookup
	expectStandardQuery(s.db, "SELECT")

	// Preload the beneficiary account insert
	s.db.ExpectBegin()
	expectStandardQuery(s.db, "INSERT")
	s.db.ExpectCommit()

	// Transaction record update
	s.db.ExpectBegin()
	s.db.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	s.db.ExpectCommit()

	// Seal the envelope using the public key
	key, err := s.certs.GetRSAKeys()
	require.NoError(err)
	msg, reject, err := envelope.Seal(payload, envelope.WithRSAPublicKey(&key.PublicKey))
	require.NoError(err)
	require.Nil(reject)

	// Start the gRPC client
	creds, err := mtls.ClientCreds("localhost", s.certs, s.chain)
	require.NoError(err)
	require.NoError(s.grpc.Connect(creds))
	defer s.grpc.Close()
	client := protocol.NewTRISANetworkClient(s.grpc.Conn)

	// Do the request
	response, err := client.Transfer(context.Background(), msg)
	require.NoError(err)
	require.NotNil(response)

	// Should get a rejection error
	require.Equal(envelope.Error, envelope.Status(response))
}

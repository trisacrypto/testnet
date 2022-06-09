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
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/envelope"
	"github.com/trisacrypto/trisa/pkg/trisa/mtls"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

const bufSize = 1024 * 1024

// rVASPTestSuite tests interactions with the rvasp servers to ensure that they return
// the expected error codes and messages.
type rVASPTestSuite struct {
	suite.Suite
	grpc  *bufconn.GRPCListener
	db    sqlmock.Sqlmock
	trisa *rvasp.TRISA
	certs *trust.Provider
	chain trust.ProviderPool
	peers *peers.Peers
	conf  *config.Config
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

	identity := &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
		Beneficiary:     &ivms101.Beneficiary{},
		BeneficiaryVasp: &ivms101.BeneficiaryVasp{
			BeneficiaryVasp: &ivms101.Person{},
		},
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

	// Preload the beneficiary account insert
	s.db.ExpectBegin()
	expectStandardQuery(s.db, "INSERT")
	s.db.ExpectCommit()

	// Preload the identity lookups
	expectStandardQuery(s.db, "SELECT")
	expectStandardQuery(s.db, "SELECT")

	// Preload the transaction insert
	s.db.ExpectBegin()
	// Account record insert
	expectStandardQuery(s.db, "INSERT")
	// Identity records insert
	expectStandardQuery(s.db, "INSERT")
	expectStandardQuery(s.db, "INSERT")
	// VASP record insert
	expectStandardQuery(s.db, "INSERT")
	// Transaction record insert
	expectStandardQuery(s.db, "INSERT")
	s.db.ExpectCommit()

	// Account record update
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

	// Preload the beneficiary account insert
	s.db.ExpectBegin()
	expectStandardQuery(s.db, "INSERT")
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
	_, err = client.Transfer(context.Background(), msg)
	require.EqualError(err, status.Errorf(codes.FailedPrecondition, "TRISA protocol error: missing beneficiary vasp identity").Error())
}

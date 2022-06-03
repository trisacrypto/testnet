package rvasp_test

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"github.com/trisacrypto/testnet/pkg/rvasp"
	"github.com/trisacrypto/testnet/pkg/rvasp/bufconn"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/envelope"
	"github.com/trisacrypto/trisa/pkg/trisa/mtls"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
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
	key   *rsa.PrivateKey
	peers *peers.Peers
	creds grpc.DialOption
}

func TestRVASP(t *testing.T) {
	suite.Run(t, new(rVASPTestSuite))
}

func (s *rVASPTestSuite) SetupSuite() {
	var err error
	require := s.Require()

	s.grpc = bufconn.New(bufSize)

	_, s.trisa, s.peers, s.db, s.key, err = rvasp.NewMock()
	require.NoError(err)

	go s.trisa.Run(s.grpc.Listener)
}

func (s *rVASPTestSuite) TearDownSuite() {
	if s.grpc != nil {
		s.grpc.Release()
	}
}

// Test that the TRISA server returns a valid envelope when a valid request is sent.
func (s *rVASPTestSuite) TestValidTransfer() {
	var err error
	require := s.Require()

	originatorAddress := "mary address"
	beneficiaryAddress := "robert address"

	// Create the request envelope
	payload := &protocol.Payload{
		SentAt: time.Now().Format(time.RFC3339),
	}

	identity := &ivms101.IdentityPayload{}
	payload.Identity, err = anypb.New(identity)
	require.NoError(err)

	transaction := &generic.Transaction{
		Originator:  originatorAddress,
		Beneficiary: beneficiaryAddress,
	}
	payload.Transaction, err = anypb.New(transaction)

	msg, reject, err := envelope.Seal(payload, envelope.WithRSAPublicKey(&s.key.PublicKey))
	require.NoError(err)
	require.Nil(reject)

	// Preload the query mocks
	s.db.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"wallet_address"}).AddRow(beneficiaryAddress))
	s.db.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"beneficiary_policy"}).AddRow("SyncRequire"))
	s.db.ExpectExec("UPDATE")

	// Identity lookps
	s.db.ExpectQuery("SELECT")
	s.db.ExpectExec("INSERT")
	s.db.ExpectQuery("SELECT")
	s.db.ExpectExec("INSERT")

	s.db.ExpectExec("INSERT")
	s.db.ExpectExec("UPDATE")

	gp := &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				VerifiedChains: [][]*x509.Certificate{{
					{
						Subject: pkix.Name{
							CommonName: "test-peer",
						},
					},
				}},
			},
		},
	}

	ctx := peer.NewContext(context.Background(), gp)

	rec, ok := peer.FromContext(ctx)
	require.True(ok)

	require.NotNil(rec.AuthInfo)
	fmt.Printf("type %T\n", rec.AuthInfo)

	//var tlsAuth credentials.TLSInfo
	if _, ok = rec.AuthInfo.(credentials.TLSInfo); !ok {
		fmt.Print("not tls")
		fmt.Printf("unexpected peer transport credentials type: %T", rec.AuthInfo)
	}

	CertPath := filepath.Join("testdata", "cert.pem")
	TrustChainPath := filepath.Join("testdata", "cert.pem")

	sz, err := trust.NewSerializer(false)
	require.NoError(err)

	certs, err := sz.ReadFile(CertPath)
	require.NoError(err)

	chain, err := sz.ReadPoolFile(TrustChainPath)
	require.NoError(err)

	// Start the gRPC client.
	creds, err := mtls.ClientCreds("localhost", certs, chain)
	require.NoError(err)
	require.NoError(s.grpc.Connect(creds))
	defer s.grpc.Close()
	client := protocol.NewTRISANetworkClient(s.grpc.Conn)

	response, err := client.Transfer(ctx, msg)
	require.NoError(err)
	require.NotNil(response)

	require.NoError(s.db.ExpectationsWereMet())

	// Decrypt the response envelope
	payload, reject, err = envelope.Open(response, envelope.WithRSAPrivateKey(s.key))
	require.NoError(err)
	require.Nil(reject)
	require.NotNil(payload)

	// Check for a valid response
	require.NotNil(payload.Identity)
	require.NotNil(payload.Transaction)
	require.NotEmpty(payload.SentAt)
	require.NotEmpty(payload.ReceivedAt)
}

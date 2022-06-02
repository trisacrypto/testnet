package rvasp_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/trisacrypto/testnet/pkg/rvasp"
	"github.com/trisacrypto/testnet/pkg/rvasp/bufconn"
)

const bufSize = 1024 * 1024

// rVASPTestSuite tests interactions with the rvasp servers to ensure that they return
// the expected error codes and messages.
type rVASPTestSuite struct {
	suite.Suite
	grpc  *bufconn.GRPCListener
	trisa *rvasp.TRISA
}

func TestRVASP(t *testing.T) {
	suite.Run(t, new(rVASPTestSuite))
}

func (s *rVASPTestSuite) SetupSuite() {
	require := s.Require()

	s.grpc = bufconn.New(bufSize)

	_, trisa, err := rvasp.NewMock()
	require.NoError(err)
	s.trisa = trisa
}

func (s *rVASPTestSuite) TearDownSuite() {
	if s.grpc != nil {
		s.grpc.Release()
	}
}

func (s *rVASPTestSuite) BeforeTest(suiteName, testName string) {
}

func (s *rVASPTestSuite) AfterTest() {
}

// Test that the TRISA server returns a valid envelope when a valid request is sent.
func (s *rVASPTestSuite) TestValidTransfer(t *testing.T) {
	require := s.Require()

}

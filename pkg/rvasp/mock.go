package rvasp

import (
	"crypto/rsa"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/trisacrypto/testnet/pkg/rvasp/config"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"github.com/trisacrypto/trisa/pkg/trust/mock"
	"google.golang.org/grpc"
	"software.sslmate.com/src/go-pkcs12"
)

// NewMock returns a mock rVASP server that can be used for testing.
func NewMock() (s *Server, t *TRISA, peers *peers.Peers, mockDB sqlmock.Sqlmock, key *rsa.PrivateKey, err error) {
	var conf *config.Config
	if conf, err = config.New(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	s = &Server{conf: conf, echan: make(chan error, 1)}
	if s.db, mockDB, err = db.NewDBMock("alice"); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	s.vasp = s.db.GetVASP()

	if s.trisa, err = NewTRISAMock(s); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return s, s.trisa, s.peers, mockDB, s.trisa.sign, nil
}

// NewTRISAMock returns a mock TRISA server that can be used for testing.
func NewTRISAMock(parent *Server) (s *TRISA, err error) {
	conf := parent.conf

	var pfxData []byte
	if pfxData, err = mock.Chain(); err != nil {
		return nil, err
	}

	var private *trust.Provider
	if private, err = trust.Decrypt(pfxData, pkcs12.DefaultPassword); err != nil {
		return nil, err
	}

	pool := trust.NewPool(private)
	parent.peers = peers.NewMock(private, pool, conf.GDS.URL)

	s = &TRISA{parent: parent}
	if s.sign, err = private.GetRSAKeys(); err != nil {
		return nil, err
	}

	s.srv = grpc.NewServer()
	protocol.RegisterTRISANetworkServer(s.srv, s)

	return s, nil
}

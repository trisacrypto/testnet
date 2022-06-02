package rvasp

import (
	"github.com/trisacrypto/testnet/pkg/rvasp/config"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"github.com/trisacrypto/trisa/pkg/trust/mock"
	"software.sslmate.com/src/go-pkcs12"
)

// NewMock returns a mock rVASP server that can be used for testing.
func NewMock() (s *Server, t *TRISA, err error) {
	var conf *config.Config
	if conf, err = config.New(); err != nil {
		return nil, nil, err
	}

	s = &Server{conf: conf, echan: make(chan error, 1)}
	if s.db, _, err = db.NewDBMock("alice"); err != nil {
		return nil, nil, err
	}
	s.vasp = s.db.GetVASP()

	if s.trisa, err = NewTRISAMock(s); err != nil {
		return nil, nil, err
	}

	return s, s.trisa, nil
}

// NewTRISAMock returns a mock TRISA server that can be used for testing.
func NewTRISAMock(parent *Server) (svc *TRISA, err error) {
	conf := parent.conf

	var pfxData []byte
	if pfxData, err = mock.Chain(); err != nil {
		return nil, err
	}

	var private *trust.Provider
	if private, err = trust.Decrypt(pfxData, pkcs12.DefaultPassword); err != nil {
		return nil, err
	}

	parent.peers = peers.NewMock(private, trust.NewPool(), conf.GDS.URL)

	svc = &TRISA{parent: parent}
	if svc.sign, err = private.GetRSAKeys(); err != nil {
		return nil, err
	}

	return svc, nil
}

package rvasp

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/trisacrypto/testnet/pkg/rvasp/config"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/mtls"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"github.com/trisacrypto/trisa/pkg/trust"
	"google.golang.org/grpc"
)

// NewTRISAMock returns a mock TRISA server that can be used for testing.
func NewTRISAMock(conf *config.Config) (s *TRISA, remotePeers *peers.Peers, mockDB sqlmock.Sqlmock, certs *trust.Provider, chain trust.ProviderPool, err error) {
	// Create the parent server
	var parent *Server
	if parent, mockDB, err = NewServerMock(conf); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// Create the mock TRISA server from the parent server
	s = &TRISA{parent: parent}

	if s.certs, s.chain, err = loadCerts(conf); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	if s.sign, err = s.certs.GetRSAKeys(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// Create a mock remote peer cache
	remotePeers = peers.NewMock(s.certs, s.chain, conf.GDS.URL)
	remotePeers.Add(&peers.PeerInfo{
		CommonName: "alice",
		Endpoint:   "gds.example.io:443",
		SigningKey: &s.sign.PublicKey,
	})
	parent.peers = remotePeers

	var creds grpc.ServerOption
	if creds, err = mtls.ServerCreds(s.certs, s.chain); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	s.srv = grpc.NewServer(creds)
	protocol.RegisterTRISANetworkServer(s.srv, s)

	return s, remotePeers, mockDB, s.certs, s.chain, nil
}

// NewServerMock returns a mock rVASP server that can be used for testing.
func NewServerMock(conf *config.Config) (s *Server, mockDB sqlmock.Sqlmock, err error) {
	s = &Server{conf: conf, echan: make(chan error, 1)}
	if s.db, mockDB, err = db.NewDBMock("alice"); err != nil {
		return nil, nil, err
	}
	s.vasp = s.db.GetVASP()
	s.updates = NewUpdateManager()
	return s, mockDB, nil
}

// Load certificates from the configuration
func loadCerts(conf *config.Config) (certs *trust.Provider, chain trust.ProviderPool, err error) {
	// Load the certificate from disk
	var sz *trust.Serializer
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, nil, err
	}

	if certs, err = sz.ReadFile(conf.CertPath); err != nil {
		return nil, nil, err
	}

	if chain, err = sz.ReadPoolFile(conf.TrustChainPath); err != nil {
		return nil, nil, err
	}

	return certs, chain, nil
}

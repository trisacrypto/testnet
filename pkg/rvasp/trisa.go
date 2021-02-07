package rvasp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/trisa/mtls"
	protocol "github.com/trisacrypto/testnet/pkg/trisa/protocol/v1alpha1"
	"github.com/trisacrypto/testnet/pkg/trust"
	"google.golang.org/grpc"
)

// Peer implements the GRPC TRISANetwork and TRISAHealth services.
type Peer struct {
	protocol.UnimplementedTRISANetworkServer
	protocol.UnimplementedTRISAHealthServer
	parent *Server
	srv    *grpc.Server
	certs  *trust.Provider
	chain  trust.ProviderPool
}

// NewPeer from a parent server.
func NewPeer(parent *Server) (peer *Peer, err error) {
	conf := parent.conf
	if conf.CertPath == "" || conf.TrustChainPath == "" {
		return nil, errors.New("certificate path or trust chain path missing")
	}

	var sz *trust.Serializer
	peer = &Peer{parent: parent}

	// Load the TRISA certificates for server-side TLS
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, err
	}

	if peer.certs, err = sz.ReadFile(conf.CertPath); err != nil {
		return nil, err
	}

	// Load the TRISA public pool from disk
	if sz, err = trust.NewSerializer(false); err != nil {
		return nil, err
	}

	if peer.chain, err = sz.ReadPoolFile(conf.TrustChainPath); err != nil {
		return nil, err
	}

	return peer, nil
}

// Serve initializes the GRPC server and returns any errors during intitialization, it
// then kicks off a go routine to handle requests. Not thread safe, should not be called
// multiple times.
func (s *Peer) Serve() (err error) {
	// Create TLS Credentials for the server
	// NOTE: the mtls package specifies TRISA-specific TLS configuration.
	var creds grpc.ServerOption
	if creds, err = mtls.ServerCreds(s.certs, s.chain); err != nil {
		return err
	}

	s.srv = grpc.NewServer(creds)
	protocol.RegisterTRISANetworkServer(s.srv, s)
	protocol.RegisterTRISAHealthServer(s.srv, s)

	var sock net.Listener
	if sock, err = net.Listen("tcp", s.parent.conf.TRISABindAddr); err != nil {
		return fmt.Errorf("trisa peer could not listen on %q", s.parent.conf.TRISABindAddr)
	}

	go func() {
		defer sock.Close()

		log.Info().
			Str("listen", s.parent.conf.TRISABindAddr).
			Msg("trisa peer server started")

		if err := s.srv.Serve(sock); err != nil {
			s.parent.echan <- err
		}
	}()

	return nil
}

// Shutdown the TRISA peer server gracefully
func (s *Peer) Shutdown() (err error) {
	log.Info().Msg("trisa peer gracefully shutting down")
	s.srv.GracefulStop()
	log.Debug().Msg("successful trisa peer shutdown")
	return nil
}

// Transfer enables a quick one-off transaction between peers.
func (s *Peer) Transfer(ctx context.Context, in *protocol.SecureEnvelope) (out *protocol.SecureEnvelope, err error) {
	return nil, nil
}

// TransferStream allows for high-throughput transactions.
func (s *Peer) TransferStream(stream protocol.TRISANetwork_TransferStreamServer) (err error) {
	return &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented the transfer stream yet",
		Retry:   false,
	}
}

// ConfirmAddress allows the rVASP to respond to proof-of-control requests.
func (s *Peer) ConfirmAddress(ctx context.Context, in *protocol.Address) (out *protocol.AddressConfirmation, err error) {
	return nil, &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented address confirmation yet",
		Retry:   false,
	}
}

// KeyExchange facilitates signing key exchange between VASPs.
func (s *Peer) KeyExchange(ctx context.Context, in *protocol.SigningKey) (out *protocol.SigningKey, err error) {
	return nil, &protocol.Error{
		Code:    protocol.Unimplemented,
		Message: "rVASP has not implemented key exchnage yet",
		Retry:   false,
	}
}

// Status returns a directory health check status as online and requests half an hour checks.
func (s *Peer) Status(ctx context.Context, in *protocol.HealthCheck) (out *protocol.ServiceState, err error) {
	log.Info().
		Uint32("attempts", in.Attempts).
		Str("last checked at", in.LastCheckedAt).
		Msg("status check")

	// Request another health check between 30 minutes and an hour from now.
	now := time.Now()
	return &protocol.ServiceState{
		Status:    protocol.ServiceState_HEALTHY,
		NotBefore: now.Add(30 * time.Minute).Format(time.RFC3339),
		NotAfter:  now.Add(1 * time.Hour).Format(time.RFC3339),
	}, nil
}

package rvasp

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"

	protocol "github.com/trisacrypto/testnet/pkg/trisa/protocol/v1alpha1"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// Peer contains cached information about connections to other members of the TRISA
// network and facilitates directory service lookups and information exchanges.
//
// TODO: move the Peer to the TRISA library
type Peer struct {
	ID                  string
	RegisteredDirectory string
	CommonName          string
	Endpoint            string
	SigningKey          *rsa.PublicKey
	client              protocol.TRISANetworkClient
	stream              protocol.TRISANetwork_TransferStreamClient
}

func (s *Server) peerFromContext(ctx context.Context) (_ *Peer, err error) {
	var (
		ok         bool
		gp         *peer.Peer
		tlsAuth    credentials.TLSInfo
		commonName string
	)

	if gp, ok = peer.FromContext(ctx); !ok {
		return nil, errors.New("no peer found in context")
	}

	if tlsAuth, ok = gp.AuthInfo.(credentials.TLSInfo); !ok {
		return nil, fmt.Errorf("unexpected peer transport credentials type: %T", gp.AuthInfo)
	}

	if len(tlsAuth.State.VerifiedChains) == 0 || len(tlsAuth.State.VerifiedChains[0]) == 0 {
		return nil, errors.New("could not verify peer certificate")
	}

	commonName = tlsAuth.State.VerifiedChains[0][0].Subject.CommonName
	if commonName == "" {
		return nil, errors.New("could not find common name on authenticated subject")
	}

	// Check if peer is already cached in memory. If not, add the new peer.
	if _, ok = s.peers[commonName]; !ok {
		s.peers[commonName] = &Peer{
			CommonName: commonName,
		}

		// TODO: Do a directory service lookup for the ID and registered ID.
	}
	return s.peers[commonName], nil
}

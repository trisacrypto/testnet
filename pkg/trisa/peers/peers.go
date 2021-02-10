package peers

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"

	protocol "github.com/trisacrypto/testnet/pkg/trisa/protocol/v1alpha1"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// Peers maps the common name from the mTLS certificate to a Peer structure.
type Peers map[string]*Peer

// Peer contains cached information about connections to other members of the TRISA
// network and facilitates directory service lookups and information exchanges.
type Peer struct {
	ID                  string
	RegisteredDirectory string
	CommonName          string
	Endpoint            string
	SigningKey          *rsa.PublicKey
	client              protocol.TRISANetworkClient
	stream              protocol.TRISANetwork_TransferStreamClient
}

// New creates a new peers cache to look up peers from context.
func New() Peers {
	return make(Peers)
}

// FromContext looks up the TLSInfo from the incoming gRPC connection to get the common
// name of the Peer from the certificate. If the Peer is already in the cache, it
// returns the peer information, otherwise it creates and caches the Peer info.
func (p Peers) FromContext(ctx context.Context) (_ *Peer, err error) {
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
	if _, ok = p[commonName]; !ok {
		p[commonName] = &Peer{
			CommonName: commonName,
		}

		// TODO: Do a directory service lookup for the ID and registered ID.
	}
	return p[commonName], nil
}

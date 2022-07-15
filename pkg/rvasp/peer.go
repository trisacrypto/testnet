package rvasp

import (
	"crypto/rsa"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
)

// TODO: The peers cache needs to be flushed periodically to prevent the situation
// where remote peers change their endpoints or certificates and the rVASPs are stuck
// using the old info.

// fetchPeer returns the peer with the common name from the cache and performs a lookup
// against the directory service if the peer does not have an endpoint.
func (s *Server) fetchPeer(commonName string) (peer *peers.Peer, err error) {
	// Retrieve or create the peer from the cache
	if peer, err = s.peers.Get(commonName); err != nil {
		log.Error().Err(err).Msg("could not create or fetch peer")
		return nil, fmt.Errorf("could not create or fetch peer: %s", err)
	}

	// Ensure that the remote peer has an endpoint to connect to
	if err = s.resolveEndpoint(peer); err != nil {
		log.Warn().Err(err).Msg("could not fetch endpoint from remote peer")
		return nil, fmt.Errorf("could not fetch endpoint from remote peer: %s", err)
	}

	return peer, nil
}

// fetchSigningKey returns the signing key for the peer, performing an endpoint lookup
// and key exchange if necessary.
func (s *Server) fetchSigningKey(peer *peers.Peer) (key *rsa.PublicKey, err error) {
	// Ensure that the remote peer has an endpoint to connect to
	if err = s.resolveEndpoint(peer); err != nil {
		log.Warn().Err(err).Msg("could not fetch endpoint from remote peer")
		return nil, fmt.Errorf("could not fetch endpoint from remote peer: %s", err)
	}

	if peer.SigningKey() == nil {
		// If no key is available, perform a key exchange with the remote peer
		if peer.ExchangeKeys(true); err != nil {
			log.Warn().Str("common_name", peer.String()).Err(err).Msg("could not exchange keys with remote peer")
			return nil, fmt.Errorf("could not exchange keys with remote peer: %s", err)
		}

		// Verify the key is now available on the peer
		if peer.SigningKey() == nil {
			log.Error().Str("common_name", peer.String()).Msg("peer has no key after key exchange")
			return nil, fmt.Errorf("peer has no key after key exchange")
		}
	}

	return peer.SigningKey(), nil
}

// resolveEndpoint ensures that the peer has an endpoint to connect to, and performs a
// lookup against the directory service to set the endpoint on the peer if necessary.
func (s *Server) resolveEndpoint(peer *peers.Peer) (err error) {
	// If the endpoint is not in the peer, do the lookup to fetch the endpoint
	if peer.Info().Endpoint == "" {
		var remote *peers.Peer
		if remote, err = s.peers.Lookup(peer.String()); err != nil {
			log.Warn().Str("peer", peer.String()).Err(err).Msg("could not lookup peer")
			return fmt.Errorf("could not lookup peer: %s", err)
		}

		if remote.Info().Endpoint == "" {
			log.Error().Str("peer", peer.String()).Msg("peer has no endpoint after lookup")
			return fmt.Errorf("peer has no endpoint after lookup")
		}
	}
	return nil
}

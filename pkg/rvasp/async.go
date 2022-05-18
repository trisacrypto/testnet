package rvasp

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	"github.com/trisacrypto/trisa/pkg/trisa/envelope"
	"github.com/trisacrypto/trisa/pkg/trisa/peers"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

// AsyncDispatcher is a go routine that periodically reads pending messages off the
// rVASP database and initiates TRISA transfers back to the originator. This allows the
// rVASPs to simulate asynchronous transactions.
func (s *TRISA) AsyncDispatcher(stop <-chan struct{}) {
	ticker := time.NewTicker(s.parent.conf.AsyncInterval)

	log.Info().Dur("Interval", s.parent.conf.AsyncInterval).Msg("asynchronous dispatcher started")

	for {
		// Wait for the next tick
		<-ticker.C

		// Retrieve all pending messages from the database
		var (
			transactions []db.Transaction
			err          error
		)
		if err = s.parent.db.LookupPending().Find(&transactions).Error; err != nil {
			log.Error().Err(err).Msg("could not lookup transactions")
			continue
		}

		now := time.Now()
		for _, tx := range transactions {
			// Verify pending transaction is old enough
			if now.Before(tx.NotBefore) {
				continue
			}

			// Verify pending transaction has not expired
			if now.After(tx.NotAfter) {
				log.Info().Uint("id", tx.ID).Time("not_after", tx.NotAfter).Msg("transaction expired")
				continue
			}

			// Acknowledge the transaction with the originator
			if err = s.acknowledgeTransaction(tx); err != nil {
				log.Error().Err(err).Uint("id", tx.ID).Msg("could not acknowledge transaction")
				continue
			}

			// Mark the transaction as completed
			tx.State = db.TransactionCompleted
			if err = s.parent.db.Save(&tx).Error; err != nil {
				log.Error().Err(err).Uint("id", tx.ID).Msg("could not save transaction")
				continue
			}
		}

		select {
		case <-stop:
			log.Info().Msg("asynchronous dispatcher received stop signal")
			return
		default:
		}
	}
}

// acknowledgeTransaction acknowledges a received transaction by initiating a transfer
// with the originator and sending back the transaction with a received_at timestamp.
func (s *TRISA) acknowledgeTransaction(tx db.Transaction) (err error) {
	// Conduct a TRISADS lookup if necessary to get the endpoint
	// TODO: Populate the originator information when we first receive the transaction
	var peer *peers.Peer
	if peer, err = s.parent.peers.Search(tx.Originator.Provider); err != nil {
		log.Error().Err(err).Msg("could not search peer from directory service")
		return fmt.Errorf("could not search peer from directory service: %s", err)
	}

	// Ensure that the local RVASP has signing keys for the remote, otherwise perform key exchange
	var signKey *rsa.PublicKey
	if signKey, err = peer.ExchangeKeys(true); err != nil {
		log.Error().Err(err).Msg("could not exchange keys with remote peer")
		return fmt.Errorf("could not exchange keys with remote peer: %s", err)
	}

	// Create the identity for the payload
	identity := &ivms101.IdentityPayload{}
	if err = protojson.Unmarshal([]byte(tx.Identity), identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity from transaction")
		return fmt.Errorf("could not unmarshal identity from transaction: %s", err)
	}

	// Create the payload
	payload := &protocol.Payload{
		SentAt:     time.Now().Format(time.RFC3339),
		ReceivedAt: time.Now().Format(time.RFC3339),
	}
	if payload.Identity, err = anypb.New(identity); err != nil {
		log.Error().Err(err).Msg("could not dump payload identity")
		return fmt.Errorf("could not dump payload identity: %s", err)
	}

	// Secure the envelope with the remote originator's signing keys
	msg, _, err := envelope.Seal(payload, envelope.WithEnvelopeID(tx.Envelope), envelope.WithRSAPublicKey(signKey))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while sealing envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	// Conduct the TRISA transfer, handle errors
	if msg, err = peer.Transfer(msg); err != nil {
		log.Error().Err(err).Msg("could not perform TRISA exchange")
		return fmt.Errorf("could not perform TRISA exchange: %s", err)
	}

	// Open the response envelope with local private keys
	payload, _, err = envelope.Open(msg, envelope.WithRSAPrivateKey(s.parent.trisa.sign))
	if err != nil {
		log.Error().Err(err).Msg("TRISA protocol error while opening envelope")
		return fmt.Errorf("TRISA protocol error: %s", err)
	}

	// Confirm that the identity payload was echoed back
	if payload.Identity == nil {
		log.Error().Msg("TRISA protocol error: did not receive identity payload")
		return fmt.Errorf("TRISA protocol error: did not receive identity payload")
	}

	return nil
}

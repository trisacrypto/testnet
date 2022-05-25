package rvasp

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
)

// AsyncDispatcher is a go routine that periodically reads pending messages off the
// rVASP database and initiates TRISA transfers back to the originator. This allows the
// rVASPs to simulate asynchronous transactions.
func (s *TRISA) AsyncDispatcher(stop <-chan struct{}) {
	ticker := time.NewTicker(s.parent.conf.AsyncInterval)

	log.Info().Dur("Interval", s.parent.conf.AsyncInterval).Msg("asynchronous dispatcher started")

	for {
		// Wait for the next tick or the stop signal
		select {
		case <-stop:
			log.Info().Msg("asynchronous dispatcher received stop signal")
			return
		case <-ticker.C:
		default:
		}

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
				tx.State = db.TransactionExpired
				if err = s.parent.db.Save(&tx).Error; err != nil {
					log.Error().Err(err).Uint("id", tx.ID).Msg("could not save expired transaction")
				}
				continue
			}

			// Acknowledge the transaction with the originator
			if err = s.acknowledgeTransaction(tx); err != nil {
				log.Error().Err(err).Uint("id", tx.ID).Msg("could not acknowledge transaction")
				tx.State = db.TransactionFailed
				if err = s.parent.db.Save(&tx).Error; err != nil {
					log.Error().Err(err).Uint("id", tx.ID).Msg("could not save failed transaction")
				}
				continue
			}

			// Mark the transaction as completed
			tx.State = db.TransactionCompleted
			if err = s.parent.db.Save(&tx).Error; err != nil {
				log.Error().Err(err).Uint("id", tx.ID).Msg("could not save completed transaction")
			}
		}
	}
}

// acknowledgeTransaction acknowledges a received transaction by initiating a transfer
// with the originator and sending back the transaction with a received_at timestamp.
func (s *TRISA) acknowledgeTransaction(tx db.Transaction) (err error) {
	// Retrieve the wallet for the beneficiary account
	var wallet *db.Wallet
	if wallet, err = tx.Wallet(s.parent.db); err != nil {
		log.Warn().Err(err).Msg("could not retrieve beneficiary wallet")
	}

	policy := wallet.Policy
	switch policy {
	case db.FullAsync:
		return s.fullAsyncTransfer(tx)
	case db.RejectedAsync:
		return s.rejectedAsyncTransfer(tx)
	default:
		return fmt.Errorf("unknown policy '%s' for wallet '%s'", policy, wallet.Address)
	}
}

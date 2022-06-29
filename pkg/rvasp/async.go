package rvasp

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
)

// AsyncHandler is a go routine that periodically reads pending messages off the
// rVASP database and initiates TRISA transfers back to the originator. This allows the
// rVASPs to simulate asynchronous transactions.
func (s *TRISA) AsyncHandler(stop <-chan struct{}) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	// Create default context and cancel to stop async handler routine before a handler is fired.
	_, cancel = context.WithCancel(context.Background())

	// Create the ticker to run aync handlers periodically
	ticker := time.NewTicker(s.parent.conf.AsyncInterval)
	log.Info().Dur("interval", s.parent.conf.AsyncInterval).Msg("asynchronous handler started")

	for {
		// Wait for the next tick or the stop signal
		select {
		case <-stop:
			log.Info().Msg("asynchronous handler received stop signal")
			cancel()
			return
		case <-ticker.C:
			cancel()
		}

		// Execute the handle async go routine with cancellation signals
		log.Debug().Msg("checking for pending transactions to handle async")
		ctx, cancel = context.WithTimeout(context.Background(), s.parent.conf.AsyncInterval)
		go s.handleAsync(ctx)
	}
}

func (s *TRISA) handleAsync(ctx context.Context) {
	// Recover from panics that occur during async handling.
	defer func() {
		if r := recover(); r != nil {
			log.WithLevel(zerolog.PanicLevel).
				Err(fmt.Errorf("%v", r)).
				Str("stack_trace", string(debug.Stack())).
				Msg("handle async has recovered from a panic")
		}
	}()

	// Retrieve all pending messages from the database
	var (
		transactions []db.Transaction
		err          error
	)
	if err = s.parent.db.LookupPending().Find(&transactions).Error; err != nil {
		log.Error().Err(err).Msg("could not lookup transactions")
		return
	}

	now := time.Now()
txloop:
	for _, tx := range transactions {
		// Verify pending transaction is old enough
		if now.Before(tx.NotBefore) {
			continue
		}

		// Verify pending transaction has not expired
		if now.After(tx.NotAfter) {
			log.Info().Uint("id", tx.ID).Time("not_after", tx.NotAfter).Msg("transaction expired")
			tx.State = pb.TransactionState_EXPIRED
			if err = s.parent.db.Save(&tx).Error; err != nil {
				log.Error().Err(err).Uint("id", tx.ID).Msg("could not save expired transaction")
			}
			continue txloop
		}

		// Acknowledge the transaction with the originator
		if err = s.acknowledgeTransaction(&tx); err != nil {
			log.Warn().Err(err).Uint("id", tx.ID).Msg("could not acknowledge transaction")
			tx.State = pb.TransactionState_FAILED
		}

		// Save the updated transaction in the database
		if err = s.parent.db.Save(&tx).Error; err != nil {
			log.Error().Err(err).Uint("id", tx.ID).Msg("could not save completed transaction")
		}
	}
}

// acknowledgeTransaction acknowledges a received transaction by initiating a transfer
// with the originator depending on the configured policy in the beneficiary wallet.
func (s *TRISA) acknowledgeTransaction(tx *db.Transaction) (err error) {
	// Retrieve the local account for the transaction
	var account *db.Account
	if account, err = tx.GetAccount(s.parent.db); err != nil {
		log.Warn().Err(err).Msg("could not retrieve beneficiary account")
		return
	}

	// Retrieve the wallet for the beneficiary account
	var wallet *db.Wallet
	if wallet, err = account.GetWallet(s.parent.db); err != nil {
		log.Warn().Err(err).Msg("could not retrieve beneficiary wallet")
		return err
	}

	policy := wallet.BeneficiaryPolicy
	switch policy {
	case db.AsyncRepair:
		return s.sendAsync(tx)
	case db.AsyncReject:
		return s.sendRejected(tx)
	default:
		return fmt.Errorf("unknown policy '%s' for wallet '%s'", policy, wallet.Address)
	}
}

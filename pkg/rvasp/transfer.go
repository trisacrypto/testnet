package rvasp

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	"github.com/trisacrypto/trisa/pkg/ivms101"
	protocol "github.com/trisacrypto/trisa/pkg/trisa/api/v1beta1"
	generic "github.com/trisacrypto/trisa/pkg/trisa/data/generic/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Create an IVMS101 identity payload for a TRISA transfer.
func (s *Server) createIdentityPayload(originatorAccount db.Account, beneficiaryAccount db.Account) (identity *ivms101.IdentityPayload, err error) {
	identity = &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
		Beneficiary:     &ivms101.Beneficiary{},
		BeneficiaryVasp: &ivms101.BeneficiaryVasp{},
	}

	// Create the originator identity
	if identity.OriginatingVasp.OriginatingVasp, err = s.vasp.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load originator vasp")
		return nil, status.Errorf(codes.Internal, "could not load originator vasp: %s", err)
	}

	identity.Originator = &ivms101.Originator{
		OriginatorPersons: make([]*ivms101.Person, 0, 1),
		AccountNumbers:    []string{originatorAccount.WalletAddress},
	}
	var originator *ivms101.Person
	if originator, err = originatorAccount.LoadIdentity(); err != nil {
		log.Error().Err(err).Msg("could not load originator identity")
		return nil, status.Errorf(codes.Internal, "could not load originator identity: %s", err)
	}
	identity.Originator.OriginatorPersons = append(identity.Originator.OriginatorPersons, originator)

	// Create the beneficiary VASP identity information if provided
	if beneficiaryAccount.VaspID != 0 {
		var beneficiaryVasp *db.VASP
		if beneficiaryVasp, err = beneficiaryAccount.GetVASP(s.db); err != nil {
			log.Error().Err(err).Msg("could not fetch beneficiary from database")
			return nil, status.Errorf(codes.Internal, "could not load fetch beneficiary from database: %s", err)
		}

		// Create the beneficiary identity
		if identity.BeneficiaryVasp.BeneficiaryVasp, err = beneficiaryVasp.LoadIdentity(); err != nil {
			log.Error().Err(err).Msg("could not load beneficiary identity")
			return nil, status.Errorf(codes.Internal, "could not load beneficiary identity: %s", err)
		}
	}

	// Create the beneficiary account information if provided
	if beneficiaryAccount.WalletAddress != "" {
		identity.Beneficiary = &ivms101.Beneficiary{
			BeneficiaryPersons: make([]*ivms101.Person, 0, 1),
			AccountNumbers:     []string{beneficiaryAccount.WalletAddress},
		}
	}

	// Create the beneficiary identity if provided
	if beneficiaryAccount.IVMS101 != "" {
		var beneficiary *ivms101.Person
		if beneficiary, err = beneficiaryAccount.LoadIdentity(); err != nil {
			log.Error().Err(err).Msg("could not load beneficiary identity")
			return nil, status.Errorf(codes.Internal, "could not load beneficiary identity: %s", err)
		}
		identity.Beneficiary.BeneficiaryPersons = append(identity.Beneficiary.BeneficiaryPersons, beneficiary)
	}

	return identity, nil
}

// Create a TRISA transfer payload from an IVMS101 identity payload and a TRISA
// transaction payload.
// TODO: Refactor to a "builder" pattern to make this easier to use.
func createTransferPayload(identity *ivms101.IdentityPayload, transaction *generic.Transaction) (payload *protocol.Payload, err error) {
	// Create the payload
	payload = &protocol.Payload{
		SentAt: time.Now().Format(time.RFC3339),
	}

	if identity != nil {
		if payload.Identity, err = anypb.New(identity); err != nil {
			log.Error().Err(err).Msg("could not dump payload identity")
			return nil, fmt.Errorf("could not dump payload identity: %s", err)
		}
	}

	if transaction != nil {
		if payload.Transaction, err = anypb.New(transaction); err != nil {
			log.Error().Err(err).Msg("could not dump payload transaction")
			return nil, fmt.Errorf("could not dump payload transaction: %s", err)
		}
	}

	return payload, nil
}

func createPendingPayload(pending *generic.Pending, identity *ivms101.IdentityPayload) (payload *protocol.Payload, err error) {
	// Create the payload
	payload = &protocol.Payload{
		SentAt:     time.Now().Format(time.RFC3339),
		ReceivedAt: time.Now().Format(time.RFC3339),
	}

	if pending == nil {
		return nil, fmt.Errorf("nil pending message supplied")
	}

	if identity == nil {
		return nil, fmt.Errorf("nil identity payload supplied")
	}

	if payload.Identity, err = anypb.New(identity); err != nil {
		log.Error().Err(err).Msg("could not dump identity payload")
		return nil, fmt.Errorf("could not dump identity payload: %s", err)
	}

	if payload.Transaction, err = anypb.New(pending); err != nil {
		log.Error().Err(err).Msg("could not dump pending message")
		return nil, fmt.Errorf("could not dump pending message: %s", err)
	}

	return payload, nil
}

// Parse a TRISA transfer payload, returning the identity payload and either the
// transaction payload or the pending message.
func parsePayload(payload *protocol.Payload, response bool) (identity *ivms101.IdentityPayload, transaction *generic.Transaction, pending *generic.Pending, parseError *protocol.Error) {
	// Verify the sent_at timestamp if this is an accepted response payload
	if payload.SentAt == "" {
		log.Warn().Msg("missing sent at timestamp")
		return nil, nil, nil, protocol.Errorf(protocol.MissingFields, "missing sent_at timestamp")
	}

	// Payload must contain an identity
	if payload.Identity == nil {
		log.Warn().Msg("payload does not contain an identity")
		return nil, nil, nil, protocol.Errorf(protocol.MissingFields, "missing identity payload")
	}

	// Payload must contain a transaction
	if payload.Transaction == nil {
		log.Warn().Msg("payload does not contain a transaction")
		return nil, nil, nil, protocol.Errorf(protocol.MissingFields, "missing transaction payload")
	}

	// Parse the identity payload
	identity = &ivms101.IdentityPayload{}
	var err error
	if err = payload.Identity.UnmarshalTo(identity); err != nil {
		log.Warn().Err(err).Msg("could not unmarshal identity")
		return nil, nil, nil, protocol.Errorf(protocol.UnparseableIdentity, "could non unmarshal identity: %s", err)
	}

	// Validate identity fields
	if identity.Originator == nil || identity.OriginatingVasp == nil || identity.BeneficiaryVasp == nil || identity.Beneficiary == nil {
		log.Warn().Msg("incomplete identity payload")
		return nil, nil, nil, protocol.Errorf(protocol.IncompleteIdentity, "incomplete identity payload")
	}

	// Parse the transaction message type
	var msgTx proto.Message
	if msgTx, err = payload.Transaction.UnmarshalNew(); err != nil {
		log.Warn().Err(err).Str("transaction_type", payload.Transaction.TypeUrl).Msg("could not unmarshal incoming transaction payload")
		return nil, nil, nil, protocol.Errorf(protocol.UnparseableTransaction, "could not unmarshal transaction payload: %s", err)
	}

	switch tx := msgTx.(type) {
	case *generic.Transaction:
		transaction = tx

		// Verify the received_at timestamp if this is an accepted response payload
		// NOTE: pending messages and intermediate messages will not contain received_at
		if response && payload.ReceivedAt == "" {
			log.Warn().Msg("missing received at timestamp")
			return nil, nil, nil, protocol.Errorf(protocol.MissingFields, "missing received_at timestamp")
		}
	case *generic.Pending:
		pending = tx
	default:
		log.Warn().Str("transaction_type", payload.Transaction.TypeUrl).Msg("could not handle incoming transaction payload")
		return nil, nil, nil, protocol.Errorf(protocol.UnparseableTransaction, "unexpected transaction payload type: %s", payload.Transaction.TypeUrl)
	}
	return identity, transaction, pending, nil
}

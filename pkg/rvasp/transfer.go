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
	"google.golang.org/protobuf/types/known/anypb"
)

// Create an IVMS101 identity payload for a TRISA transfer.
func (s *Server) createIdentityPayload(originatorAccount db.Account, beneficiaryAddress string) (identity *ivms101.IdentityPayload, err error) {
	identity = &ivms101.IdentityPayload{
		Originator:      &ivms101.Originator{},
		OriginatingVasp: &ivms101.OriginatingVasp{},
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

	// Create the beneficiary identity if the address was provided
	if beneficiaryAddress != "" {
		identity.BeneficiaryVasp = &ivms101.BeneficiaryVasp{}
		identity.BeneficiaryVasp.BeneficiaryVasp = &ivms101.Person{}
		identity.Beneficiary = &ivms101.Beneficiary{
			BeneficiaryPersons: make([]*ivms101.Person, 0, 1),
			AccountNumbers:     []string{beneficiaryAddress},
		}
		identity.Beneficiary.BeneficiaryPersons = append(identity.Beneficiary.BeneficiaryPersons, &ivms101.Person{})
	}

	return identity, nil
}

// Create a TRISA transfer payload from an IVMS101 identity payload and a TRISA
// transaction payload.
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

func createPendingPayload(pending *generic.Pending) (payload *protocol.Payload, err error) {
	// Create the payload
	payload = &protocol.Payload{
		SentAt: time.Now().Format(time.RFC3339),
	}

	if pending == nil {
		return nil, fmt.Errorf("nil pending message supplied")
	}

	if payload.Identity, err = anypb.New(&ivms101.IdentityPayload{}); err != nil {
		return nil, fmt.Errorf("could not dump payload identity: %s", err)
	}

	if payload.Transaction, err = anypb.New(pending); err != nil {
		log.Error().Err(err).Msg("could not dump pending message")
		return nil, fmt.Errorf("could not dump pending message: %s", err)
	}

	return payload, nil
}

// Parse the identity payload out of a TRISA transfer payload.
func parseIdentityPayload(payload *protocol.Payload) (identity *ivms101.IdentityPayload, err error) {
	if payload.Identity == nil {
		return nil, fmt.Errorf("missing identity payload")
	}

	// Verify the returned payload type
	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unsupported identity type")
		return nil, fmt.Errorf("unsupported identity type: %s", payload.Identity.TypeUrl)
	}

	// Parse the identity payload
	identity = &ivms101.IdentityPayload{}
	if err = payload.Identity.UnmarshalTo(identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity")
		return nil, fmt.Errorf("could non unmarshal identity: %s", err)
	}

	return identity, nil
}

// Parse the transaction payload out of a TRISA transfer payload.
func parseTransactionPayload(payload *protocol.Payload) (transaction *generic.Transaction, err error) {
	if payload.Transaction == nil {
		return nil, fmt.Errorf("missing transaction payload")
	}

	// Verify the returned payload type
	if payload.Transaction.TypeUrl != "type.googleapis.com/trisa.data.generic.v1beta1.Transaction" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported transaction type")
		return nil, fmt.Errorf("unsupported transaction type: %s", payload.Transaction.TypeUrl)
	}

	// Parse the transaction payload
	transaction = &generic.Transaction{}
	if err = payload.Transaction.UnmarshalTo(transaction); err != nil {
		log.Error().Err(err).Msg("could not unmarshal transaction")
		return nil, fmt.Errorf("could not unmarshal transaction: %s", err)
	}

	return transaction, nil
}

// Parse a pending message out of a TRISA transfer payload.
func parsePendingMessage(payload *protocol.Payload) (pending *generic.Pending, err error) {
	if payload.Transaction == nil {
		return nil, fmt.Errorf("missing transaction payload")
	}

	// Verify the returned payload type
	if payload.Transaction.TypeUrl != "type.googleapis.com/trisa.data.generic.v1beta1.Pending" {
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unsupported pending type")
		return nil, fmt.Errorf("unsupported pending type: %s", payload.Transaction.TypeUrl)
	}

	// Parse the pending payload
	pending = &generic.Pending{}
	if err = payload.Transaction.UnmarshalTo(pending); err != nil {
		log.Error().Err(err).Msg("could not unmarshal pending")
		return nil, fmt.Errorf("could not unmarshal pending: %s", err)
	}

	// Timestamps should be filled in
	if pending.ReplyNotBefore == "" || pending.ReplyNotAfter == "" {
		log.Error().Msg("missing ReplyNotBefore or ReplyNotAfter timestamp in pending message")
		return nil, fmt.Errorf("missing ReplyNotBefore or ReplyNotAfter timestamp in pending message")
	}

	return pending, nil
}

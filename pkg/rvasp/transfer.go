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
func parsePayload(payload *protocol.Payload, response bool) (identity *ivms101.IdentityPayload, transaction *generic.Transaction, pending *generic.Pending, err error) {
	// Verify the received_at timestamp if this is a response payload
	if response && payload.ReceivedAt == "" {
		log.Warn().Msg("missing received at timestamp")
		return nil, nil, nil, fmt.Errorf("missing received_at timestamp")
	}

	// Payload must contain an identity
	if payload.Identity == nil {
		log.Warn().Msg("payload does not contain an identity")
		return nil, nil, nil, fmt.Errorf("missing identity payload")
	}

	// Parse the identity payload
	if payload.Identity.TypeUrl != "type.googleapis.com/ivms101.IdentityPayload" {
		log.Warn().Str("type", payload.Identity.TypeUrl).Msg("unexpected identity payload type")
		return nil, nil, nil, fmt.Errorf("unexpected identity payload type: %s", payload.Identity.TypeUrl)
	}

	identity = &ivms101.IdentityPayload{}
	if err = payload.Identity.UnmarshalTo(identity); err != nil {
		log.Error().Err(err).Msg("could not unmarshal identity")
		return nil, nil, nil, fmt.Errorf("could non unmarshal identity: %s", err)
	}

	// Parse the message type
	switch payload.Transaction.TypeUrl {
	case "type.googleapis.com/trisa.data.generic.v1beta1.Transaction":
		transaction = &generic.Transaction{}
		if err = payload.Transaction.UnmarshalTo(transaction); err != nil {
			log.Error().Err(err).Msg("could not unmarshal transaction")
			return nil, nil, nil, fmt.Errorf("could not unmarshal transaction: %s", err)
		}
	case "type.googleapis.com/trisa.data.generic.v1beta1.Pending":
		pending = &generic.Pending{}
		if err = payload.Transaction.UnmarshalTo(pending); err != nil {
			log.Error().Err(err).Msg("could not unmarshal pending message")
			return nil, nil, nil, fmt.Errorf("could not unmarshal pending message: %s", err)
		}
	default:
		log.Warn().Str("type", payload.Transaction.TypeUrl).Msg("unexpected transaction payload type")
		return nil, nil, nil, fmt.Errorf("unexpected transaction payload type: %s", payload.Transaction.TypeUrl)
	}
	return identity, transaction, pending, nil
}

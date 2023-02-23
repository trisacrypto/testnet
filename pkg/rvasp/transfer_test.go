package rvasp_test

import (
	"github.com/trisacrypto/testnet/pkg/rvasp"
	"github.com/trisacrypto/trisa/pkg/ivms101"
)

func (s *rVASPTestSuite) TestValidateIdentityPayload() {
	var err error
	require := s.Require()

	//
	err = rvasp.ValidateIdentityPayload(nil, false)
	require.EqualError(err, "trisa rejection [INTERNAL_ERROR]: identity payload is nil")

	//
	req := &ivms101.IdentityPayload{Originator: nil}
	err = rvasp.ValidateIdentityPayload(req, false)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing originator")

	//
	req.Originator = &ivms101.Originator{}
	err = rvasp.ValidateIdentityPayload(req, false)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing originating vasp")

	//
	req.OriginatingVasp = &ivms101.OriginatingVasp{}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary")

	//
	req.Beneficiary = &ivms101.Beneficiary{BeneficiaryPersons: nil}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary person")

	//
	req.Beneficiary.BeneficiaryPersons = make([]*ivms101.Person, 1)
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: unknown beneficiary person type: <nil>")

	//
	req.Beneficiary.BeneficiaryPersons[0] = &ivms101.Person{
		Person: &ivms101.Person_NaturalPerson{
			NaturalPerson: &ivms101.NaturalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary natural person validation error: one or more natural person name identifiers is required")

	//
	req.Beneficiary.BeneficiaryPersons[0] = &ivms101.Person{
		Person: &ivms101.Person_LegalPerson{
			LegalPerson: &ivms101.LegalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary legal person validation error: one or more legal person name identifiers is required")

	//
	req.Beneficiary.BeneficiaryPersons[0] = &ivms101.Person{
		Person: &ivms101.Person_LegalPerson{
			LegalPerson: &ivms101.LegalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary legal person validation error: one or more legal person name identifiers is required")

	//
	req.Beneficiary.BeneficiaryPersons[0].GetLegalPerson().Name = &ivms101.LegalPersonName{
		NameIdentifiers: []*ivms101.LegalPersonNameId{
			{
				LegalPersonName:               "LegalPersonName",
				LegalPersonNameIdentifierType: ivms101.LegalPersonNameTypeCode_LEGAL_PERSON_NAME_TYPE_CODE_LEGL,
			},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary account number")

	//
	req.Beneficiary.AccountNumbers = []string{"AccountNumber"}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary vasp")

	//
	req.BeneficiaryVasp = &ivms101.BeneficiaryVasp{}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary vasp entity")

	//
	req.BeneficiaryVasp.BeneficiaryVasp = &ivms101.Person{}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: unknown beneficiary person type: <nil>")
}

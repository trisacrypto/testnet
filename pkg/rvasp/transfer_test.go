package rvasp_test

import (
	"github.com/trisacrypto/testnet/pkg/rvasp"
	"github.com/trisacrypto/trisa/pkg/ivms101"
)

func (s *rVASPTestSuite) TestValidateIdentityPayload() {
	var err error
	require := s.Require()

	// Should return an error if the identity payload is nil
	err = rvasp.ValidateIdentityPayload(nil, false)
	require.EqualError(err, "trisa rejection [INTERNAL_ERROR]: identity payload is nil")

	// Should return an error if the originator is nil
	req := &ivms101.IdentityPayload{}
	err = rvasp.ValidateIdentityPayload(req, false)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing originator")

	// Should return an error if the originating vasp is nil
	req.Originator = &ivms101.Originator{}
	err = rvasp.ValidateIdentityPayload(req, false)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing originating vasp")

	// Should return an error if the beneficiary is nil
	req.OriginatingVasp = &ivms101.OriginatingVasp{}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary")

	// Should return an error if the beneficiary person is nil
	req.Beneficiary = &ivms101.Beneficiary{BeneficiaryPersons: nil}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary person")

	// Should return an error with a beneficiary person type other than natural or legal
	req.Beneficiary.BeneficiaryPersons = make([]*ivms101.Person, 1)
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: unknown beneficiary person type: <nil>")

	// Should return an error if the beneficiary natural person is incomplete
	req.Beneficiary.BeneficiaryPersons[0] = &ivms101.Person{
		Person: &ivms101.Person_NaturalPerson{
			NaturalPerson: &ivms101.NaturalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary natural person validation error: one or more natural person name identifiers is required")

	// Should return an error if the beneficiary legal person is incomplete
	req.Beneficiary.BeneficiaryPersons[0] = &ivms101.Person{
		Person: &ivms101.Person_LegalPerson{
			LegalPerson: &ivms101.LegalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary legal person validation error: one or more legal person name identifiers is required")

	// Should return an error if there are no beneficiary account numbers
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

	// Should return an error if the beneficiary vasp is nil
	req.Beneficiary.AccountNumbers = []string{"AccountNumber"}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary vasp")

	// Should return an error if the beneficiary vasp entity is nil
	req.BeneficiaryVasp = &ivms101.BeneficiaryVasp{}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [INCOMPLETE_IDENTITY]: missing beneficiary vasp entity")

	// Should return an error with a beneficiary vasp person type other than natural or legal
	req.BeneficiaryVasp.BeneficiaryVasp = &ivms101.Person{}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: unknown beneficiary vasp person type: <nil>")

	// Should return an error if the beneficiary vasp natural person is incomplete
	req.BeneficiaryVasp.BeneficiaryVasp = &ivms101.Person{
		Person: &ivms101.Person_NaturalPerson{
			NaturalPerson: &ivms101.NaturalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary vasp natural person validation error: one or more natural person name identifiers is required")

	// Should return an error if the beneficiary vasp legal person is incomplete
	req.BeneficiaryVasp.BeneficiaryVasp = &ivms101.Person{
		Person: &ivms101.Person_LegalPerson{
			LegalPerson: &ivms101.LegalPerson{},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.EqualError(err, "trisa rejection [VALIDATION_ERROR]: beneficiary vasp legal person validation error: one or more legal person name identifiers is required")

	// Happy path
	req.BeneficiaryVasp.BeneficiaryVasp.GetLegalPerson().Name = &ivms101.LegalPersonName{
		NameIdentifiers: []*ivms101.LegalPersonNameId{
			{
				LegalPersonName:               "LegalPersonName",
				LegalPersonNameIdentifierType: ivms101.LegalPersonNameTypeCode_LEGAL_PERSON_NAME_TYPE_CODE_LEGL,
			},
		},
	}
	err = rvasp.ValidateIdentityPayload(req, true)
	require.Nil(err)
}

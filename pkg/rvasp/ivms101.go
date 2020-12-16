package rvasp

// Natural person name type codes
const (
	NaturalPersonAlias  = 0
	NaturalPersonBirth  = 1
	NaturalPersonMaiden = 2
	NaturalPersonLegal  = 3
	NaturalPersonMisc   = 4
)

// Legal person name type codes
const (
	LegalPersonLegal   = 0
	LegalPersonShort   = 1
	LegalPersonTrading = 2
)

// Address type codes
const (
	AddressTypeHome       = 0
	AddressTypeBusiness   = 1
	AddressTypeGeographic = 2
)

// National identifier type codes
const (
	NationalIdentifierARNU = 0 // Alien registration number
	NationalIdentifierCCPT = 1 // Passport number
	NationalIdentifierRAID = 2 // Registration authorrity identifier
	NationalIdentifierDRLC = 3 // Driver license number
	NationalIdentifierFIIN = 4 // Foreign investment identity number
	NationalIdentifierTXID = 5 // Tax identification number
	NationalIdentifierSOCS = 6 // Social security number
	NationalIdentifierIDCD = 7 // Identity card number
	NationalIdentifierLEIX = 8 // Legal entity identifier (LEI)
	NationalIdentifierMISC = 9 // Unspecified
)

// IVMS101Person is a top level representation of IVMS101 data that can be serialized or
// deserialized into JSON for database storage. This data can also be extracted from a
// protocol buffer representation sent between VASPs.
type IVMS101Person struct {
	NaturalPerson struct {
		Name struct {
			NameIdentifiers         []*IVMS101NameIdentifier `json:"name_identifiers"`
			LocalNameIdentifiers    []*IVMS101NameIdentifier `json:"local_name_identifiers"`
			PhoneticNameIdentifiers []*IVMS101NameIdentifier `json:"phonetic_name_identifiers"`
		} `json:"name,omitempty"`
		Address                []*IVMS101Address              `json:"geographic_addresses"`
		NationalIdentification *IVMS101NationalIdentification `json:"national_identification"`
		CustomerIdentification string                         `json:"customer_identification"`
		DateAndPlaceOfBirth    struct {
			DateOfBirth  string `json:"date_of_birth"`
			PlaceOfBirth string `json:"place_of_birth"`
		} `json:"date_and_place_of_birth"`
		CountryOfResidence string `json:"country_of_residence"`
	} `json:"natural_person"`
	LegalPerson struct {
		Name struct {
			NameIdentifiers         []*IVMS101LegalNameIdentifier `json:"name_identifiers"`
			LocalNameIdentifiers    []*IVMS101LegalNameIdentifier `json:"local_name_identifiers"`
			PhoneticNameIdentifiers []*IVMS101LegalNameIdentifier `json:"phonetic_name_identifiers"`
		} `json:"name"`
		Address                []*IVMS101Address              `json:"geographic_addresses"`
		CustomerNumber         string                         `json:"customer_number"`
		NationalIdentification *IVMS101NationalIdentification `json:"national_identification"`
		CountryOfRegistration  string                         `json:"country_of_registration"`
	} `json:"legal_person,omitempty"`
}

// IVMS101NameIdentifier is used to collect name information with type codes.
type IVMS101NameIdentifier struct {
	PrimaryIdentifier   string `json:"primary_identifier"`
	SecondaryIdentifier string `json:"secondary_identifier"`
	NameIdentifierType  int    `json:"name_identifier_type"`
}

// IVMS101LegalNameIdentifier is used to collect single legal names with type codes.
type IVMS101LegalNameIdentifier struct {
	LegalPersonName     string `json:"legal_person_name"`
	LegalPersonNameType int    `json:"legal_person_name_identifier_type"`
}

// IVMS101Address is used to specify a geographic address in a country.
type IVMS101Address struct {
	AddressType        int      `json:"address_type"`
	Department         string   `json:"department,omitempty"`
	SubDepartment      string   `json:"sub_department,omitempty"`
	StreetName         string   `json:"street_name"`
	BuildingNumber     string   `json:"building_number,omitempty"`
	BuildingName       string   `json:"building_name,omitempty"`
	Floor              string   `json:"floor,omitempty"`
	PostBox            string   `json:"post_box,omitempty"`
	Room               string   `json:"room,omitempty"`
	PostCode           string   `json:"post_code,omitempty"`
	TownName           string   `json:"town_name,omitempty"`
	TownLocationName   string   `json:"town_location_name,omitempty"`
	DistrictName       string   `json:"district_name,omitempty"`
	CountrySubDivision string   `json:"country_sub_division,omitempty"`
	AddressLine        []string `json:"address_line"`
	Country            string   `json:"country"`
}

// IVMS101NationalIdentification is an identifier issued by an appropriate issuing authority
type IVMS101NationalIdentification struct {
	NationalIdentifier     string `json:"national_identifier"`
	NationalIdentifierType int    `json:"national_identifier_type"`
	CountryOfIssue         string `json:"country_of_issue"`
	RegistrationAuthority  string `json:"registration_authority"`
}

---
title: "VASP Integration"
date: 2021-04-23T01:35:35-04:00
description: "Describes how to integrate the TRISA protocol in the TestNet"
weight: 10
---

# TRISA Integration Overview

1. Register with a TRISA Directory Service
2. Implement the TRISA Network protocol
3. Implement the TRISA Health protocol

## VASP Directory Service Registration

### Registration Overview

Before you can integrate the TRISA protocol into your VASP software, you must register with a TRISA Directory Service (DS).  The TRISA DS provides public key and TRISA remote peer connection information for registered VASPs.

Once you have registered with the TRISA DS, you will receive a KYV certificate.  The public key in the KYV certificate will be made available to other VASPs via the TRISA DS.

When registering with the DS, you will need to provide the `address:port` endpoint where your VASP implements the TRISA Network service. This address will be registered with the DS and utilized by other VASPs when your VASP is identified as the beneficiary VASP.

For integration purposes, a hosted TestNet TRISA DS instance is available for testing.  The registration process is streamlined in the TestNet to facilitate quick integration.  It is recommended to start the production DS registration while integrating with the TestNet.


### Directory Service Registration

To start registration with the TRISA DS, you will need to clone the repository at:

[https://github.com/trisacrypto/testnet](https://github.com/trisacrypto/testnet)

After compiling the go protocol buffers per the documentation, you can run the following command to start the registration process:

```
$ go run ./cmd/trisads register <json file>
```

The JSON file should include the registration information for your VASP in the TRIXO questionnaire format.  Below is a sample JSON-like form representing the TRIXO questionnaire, which has been annotated with comments to explain expected values:

```
{
    // Required
    // Travel Rule Implementation Endpoint - where other TRISA peers should connect.
    // This should be an addr:port combination
    // e.g. api.alice.vaspbot.net:443
    "trisa_endpoint": "___",

    // A business website, e.g. https://mybusiness.com
    "website": "___",

    // Business Category options are:
    //     0 : UNKNOWN_ENTITY
    //     1 : PRIVATE_ORGANIZATION
    //     2 : GOVERNMENT_ENTITY
    //     3 : BUSINESS_ENTITY
    //     4 : NON_COMMERCIAL_ENTITY
    "business_category": ___,

    // Should be a date in YYYY-MM-DD format
    "established_on": "___",

    // The technical, legal, billing, and/or administrative contacts for the VASP.
    // NOTE: At least one is required for registration in a TRISA directory.
    // Contact information should be kept private in the directory service
    // and only used for email communication or verification.
    "contacts": {

        "technical": {
            // Name is required to identify and address the contact
            "name": "___",

            // An email address is required for all contacts
            "email": "___",

            // Phone number is optional, but it is strongly suggested
            "phone": "___",

            // Optional KYC data if required for the directory service contact.
            // Refers to a uniquely distinguishable individual; one single person
            "person": {

                // The distinct words used as identification for an individual
                "name": {

                    // Full name separated into primary and secondary identifier
                    // At least one occurrence of naturalPersonNameID must have the
                    // value ‘LEGL’ specified in naturalPersonNameIdentifierType.
                    "name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],
                    // The name and type of name by which the person is known,
                    // using local characters.
                    "local_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],
                    // The phonetic pronunciation of the name by which the person
                    // is known.
                    "phonetic_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ]
                },

                // The particulars of the person's location
                // Must be a valid mailing address
                // There must be at least one occurrence of either:
                //     - element addressLine
                //     or
                //     - streetName + buildingName
                //     or
                //     - streetName + buildingNumber
                "geographic_addresses" : [
                    {
                        // Identifies the nature of the address.
                        // Options include:
                        //     0 : Residential: Address is the home address.
                        //     1 : Business: Address is the business address.
                        //     2 : Geographic: Unspecified physical address
                        //     suitable for identification.
                        "address_type": ___,

                        // Include if relevant
                        // Identification of a division of a large organisation
                        // or building.
                        "department": "___",

                        // Include if relevant
                        // Identification of a sub-division of a large organisation
                        // or building.
                        "sub_department": "___",

                        // REQUIRED IF complete address not provided in addressLine
                        // field below.
                        // Name of a street or thoroughfare.
                        "street_name": "___",

                        // REQUIRED IF complete address not provided in addressLine
                        / field below.
                        // Number that identifies position of a building on street.
                        "building_number": "___",

                        // REQUIRED IF complete address not provided in addressLine
                        // field below.
                        // Name of the building or house.
                        "building_name": "___",

                        // Include if relevant
                        // Floor or storey within a building.
                        "floor": "___",

                        // Include if relevant
                        // Numbered box in a post office
                        // assigned to a person or organisation,
                        // where letters are kept until called for.
                        "post_box": "___",

                        // Include if relevant
                        // Building room number.
                        "room": "___",

                        // Include if relevant
                        // Identifier consisting of a group of letters and/or numbers
                        // added to a postal address to assist the sorting of mail.
                        "post_code": "___",

                        // Include if relevant
                        // Name of a built-up area with defined boundaries and local
                        // government.
                        "town_name": "___",

                        // Include if relevant
                        // Specific location name within the town.
                        "town_location_name": "___",

                        // Include if relevant
                        // Identifies a subdivision within a country subdivision.
                        "district_name": "___",

                        // Include if relevant
                        // Identifies a subdivision of a country for example, state,
                        // region, province, départment or county.
                        "country_sub_division": "___",

                        // REQUIRED IF streetName + buildingName/buildingNumber not
                        // provided
                        // Up to seven (7) lines may be provided.
                        "address_line": "___",

                        // REQUIRED
                        // The value used for the field country must be present on
                        // the ISO-3166-1 alpha-2 codes or the value XX.
                        "country": "___"
                    }
                ],

                // A legal person must have one of these nationalIdentifierTypes:
                //     ‘RAID’
                //     ‘MISC’
                //     ‘LEIX’
                //     ‘TXID’

                // If the value for nationalIdentifierType is ‘LEIX’,
                // nationalIdentifier must be a valid LEI.

                // If the value for nationalIdentifierType is not ‘LEIX’,
                // RegistrationAuthority MUST be populated.

                // A legal person must NOT have a value for countryOfIssue.
                "national_identification" : {
                    // Required: identifier issued by appropriate issuing authority
                    "national_identifier": "___",

                    // Required: type of identifier specified.
                    // Options include:
                    //     0 : Alien registration number, assigned by a govt agency.
                    //     1 : Passport number, assigned by a passport authority.
                    //     2 : Registration authority identifier.
                    //     3 : Driver license number, assigned to a driver's license.
                    //     4 : Foreign investment id number, assigned to foreign
                    //     investor.
                    //     5 : Tax identification number, assigned by tax authority
                    //     to entity.
                    //     6 : Social security number, assigned by a social security
                    //     agency.
                    //     7 : Identity card number, assigned by a national authority.
                    //     8 : Legal Entity Identifier, assigned in accordance with
                    //     ISO 17442
                    //     Note: The LEI is a 20-character, alpha-numeric code that
                    //     enables clear and unique identification of legal entities
                    //     participating in financial transactions.
                    //     9 : Unspecified, a national identifier which may be known
                    //     but which cannot otherwise be categorized or the category
                    //     of which the sender is unable to determine.
                    "national_identifier_type": ___,

                    // Optional: Country of the issuing authority.
                    "country_of_issue": "___",

                    // Optional: A code specifying the registration authority.
                    "registration_authority": "___"
                },

                "customer_identification": "___",

                "date_and_place_of_birth": {
                    // Definition: Date on which a person is born.
                    // Definition: A point in time, represented as a day within the
                    // calendar year.
                    // Compliant with ISO 8601.
                    // Type: Text
                    // Format: YYYY-MM-DD
                    "date_of_birth" : "___",

                    // Definition: The town and/or the city and/or the suburb
                    // and/or the country subdivision and/or the country where
                    // the person was born.
                    "place_of_birth" : "___"
                },

                // Definition: country in which a person resides (the place of a
                // person's home).
                // The value used for the field country must be present on the
                // ISO-3166-1 alpha-2 codes or the value XX.
                "country_of_residence": "___"
            }
        },
        "legal": {
            // Name is required to identify and address the contact
            "name": "___",

            // An email address is required for all contacts
            "email": "___",

            // Phone number is optional, but it is strongly suggested
            "phone": "___",

            // Optional KYC data if required for the directory service contact.
            // Refers to a uniquely distinguishable individual; one single person
            "person": {

                // The distinct words used as identification for an individual
                "name": {

                    // Full name separated into primary and secondary identifier
                    // At least one occurrence of naturalPersonNameID must have the
                    // value ‘LEGL’ specified in naturalPersonNameIdentifierType.
                    "name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],

                    // The name and type of name by which the person is known,
                    // using local characters.
                    "local_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],

                    // The phonetic pronunciation of the name by which the person
                    // is known.
                    "phonetic_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ]
                },
                "geographic_addresses": [],
                "national_identification": {},
                "customer_identification": "___",
                "date_and_place_of_birth": {},
                "country_of_residence": "___"
            }
        },
        "billing": {
            // Name is required to identify and address the contact
            "name": "___",

            // An email address is required for all contacts
            "email": "___",

            // Phone number is optional, but it is strongly suggested
            "phone": "___",

            // Optional KYC data if required for the directory service contact.
            // Refers to a uniquely distinguishable individual; one single person
            "person": {

                // The distinct words used as identification for an individual
                "name": {

                    // Full name separated into primary and secondary identifier
                    // At least one occurrence of naturalPersonNameID must have the
                    // value ‘LEGL’ specified in naturalPersonNameIdentifierType.
                    "name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],

                    // The name and type of name by which the person is known,
                    // using local characters.
                    "local_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],

                    // The phonetic pronunciation of the name by which the person
                    // is known.
                    "phonetic_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ]
                },
                "geographic_addresses": [],
                "national_identification": {},
                "customer_identification": "___",
                "date_and_place_of_birth": {},
                "country_of_residence": "___"
            }
        },
        "administrative": {
            // Name is required to identify and address the contact
            "name": "___",

            // An email address is required for all contacts
            "email": "___",

            // Phone number is optional, but it is strongly suggested
            "phone": "___",

            // Optional KYC data if required for the directory service contact.
            // Refers to a uniquely distinguishable individual; one single person
            "person": {

                // The distinct words used as identification for an individual
                "name": {

                    // Full name separated into primary and secondary identifier
                    // At least one occurrence of naturalPersonNameID must have the
                    // value ‘LEGL’ specified in naturalPersonNameIdentifierType.
                    "name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],
                    // The name and type of name by which the person is known,
                    // using local characters.
                    "local_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ],
                    // The phonetic pronunciation of the name by which the person
                    // is known.
                    "phonetic_name_identifiers": [
                        {
                            // Required
                            // This may be the family name, the maiden name or the
                            // married name, the main name, the surname, and in some
                            // cases, the entire name where the natural person’s name
                            // cannot be divided into two parts, or where the sender
                            // is unable to divide the person’s name into two parts.
                            "primary_identifier": "___",

                            // Optional
                            // These may be the forenames, familiar names, given
                            // names, initials, prefixes, suffixes or Roman
                            // numerals (where considered to be legally part of
                            // the name) or any other secondary names.
                            "secondary_identifier": "___",

                            // Required
                            // The nature of the name specified.
                            // Options include:
                            //     0 : Alias name
                            //     1 : Name at birth
                            //     2 : Maiden name
                            //     3 : Legal name
                            //     4 : Unspecified (doesn't fit into other categories)
                            "name_identifier_type": "___"
                        }
                    ]
                },
                "geographic_addresses": [],
                "national_identification": {},
                "customer_identification": "___",
                "date_and_place_of_birth": {},
                "country_of_residence": "___"
            }
        }
    },

    // The legal entity IVMS 101 data for VASP KYC information exchange.
    // This is the IVMS 101 data that should be exchanged in the TRISA P2P protocol
    // as the Originator, Intermediate, or Beneficiary VASP fields.
    // A complete and valid identity record with country of registration is required.
    "entity": {

        // The "name" field can hold multiple legal entity names, but at least one
        // must be the official name under which the organisation is legally
        // registered to do business.
        "name": {

            // The names and type of name by which the legal person is known.
            "name_identifiers": [
                {
                    // Required: Name by which the legal person is known (one or more)
                    "legal_person_name": "___",

                    // Required: The nature of the name specified.
                    // Legal name: Official name organisation is registered under.
                    "legal_person_name_identifier_type": 0
                },
                {
                    // Optional: Another name by which the legal entity is known.
                    "legal_person_name": "___",

                    // If another legal name is provided, define the type of name
                    // Options include:
                    //     0 : Legal name (official name)
                    //     1 : Short name (nickname)
                    //     2 : Trading name, other name used for commercial purposes.
                    "legal_person_name_identifier_type": ___
                }
            ],

            // The name and type of name by which the legal entity is known,
            // using local characters.
            "local_name_identifiers": [
                {
                    // Optional: Local Name (zero or more)
                    "legal_person_name": "___",

                    // If local name is provided, define what type of name it is
                    // Options include:
                    //     0 : Legal name (official name)
                    //     1 : Short name (nickname)
                    //     2 : Trading name, other name used for commercial purposes.
                    "legal_person_name_identifier_type": ___
                }
            ],

            // The phonetic pronunciation of the name by which the legal entity
            // is known.
            "phonetic_name_identifiers": [
                {
                    // Optional: Phonetic Name (zero or more)
                    "legal_person_name": "___",

                    // If phonetic name is provided, define what type of name it is
                    // Options include:
                    //     0 : Legal name (official name)
                    //     1 : Short name (nickname)
                    //     2 : Trading name, other name used for commercial purposes.
                    "legal_person_name_identifier_type": ___
                }
            ]
        },
        // The particulars of the legal entity's location
        // Must be a valid mailing address
        // There must be at least one occurrence of either:
        //     - element addressLine
        //     or
        //     - streetName + buildingName
        //     or
        //     - streetName + buildingNumber
        "geographic_addresses": [
            {
                // Identifies the nature of the address.
                // Options include:
                //     0 : Residential: Address is the home address.
                //     1 : Business: Address is the business address.
                //     2 : Geographic: Unspecified physical address suitable for
                //     identification.
                "address_type": ___,

                // Include if relevant
                // Identification of a division of a large organisation or building.
                "department": "___",

                // Include if relevant
                // Identification of a sub-division of a large organisation or
                // building.
                "sub_department": "___",

                // REQUIRED IF complete address not provided in addressLine
                // field below.
                // Name of a street or thoroughfare.
                "street_name": "___",

                // REQUIRED IF complete address not provided in addressLine
                // field below.
                // Number that identifies the position of a building on a street.
                "building_number": "___",

                // REQUIRED IF complete address not provided in addressLine
                // field below.
                // Name of the building or house.
                "building_name": "___",

                // Include if relevant
                // Floor or storey within a building.
                "floor": "___",

                // Include if relevant
                // Numbered box in a post office, assigned to a person or
                // organisation, where letters are kept until called for.
                "post_box": "___",

                // Include if relevant
                // Building room number.
                "room": "___",

                // Include if relevant
                // Identifier consisting of a group of letters and/or number
                // added to a postal address to assist the sorting of mail.
                "post_code": "___",

                // Include if relevant
                // Name of a built-up area with defined boundaries and local
                // government.
                "town_name": "___",

                // Include if relevant
                // Specific location name within the town.
                "town_location_name": "___",

                // Include if relevant
                // Identifies a subdivision within a country subdivision.
                "district_name": "___",

                // Include if relevant
                // Identifies a subdivision of a country for example, state,
                // region, province, départment or county.
                "country_sub_division": "___",

                // REQUIRED IF  streetName + buildingName/buildingNumber not
                // provided
                // Up to seven (7) lines may be provided.
                "address_line": "___",

                // REQUIRED
                // The value used for the field country must be present on the
                // ISO-3166-1 alpha-2 codes or the value XX.
                "country": "___"

            }
        ],

        // Definition: The unique identification number applied by the VASP to
        // customer.
        // NOTE The specification has a descrepency in that 5.2.9.3.3 specifies
        // an element name as "customerNumber", while the table in 5.2.9.1 calls
        // that element "customerIdentification"
        "customer_number" : "___",

        // National Identifier (REQUIRED)

        // A legal person must have one nationalIdentifierType of the following:
        //     ‘RAID’
        //     ‘MISC’
        //     ‘LEIX’
        //     ‘TXID’

        // If the value for nationalIdentifierType is ‘LEIX’,
        // nationalIdentifier must be a valid LEI.

        // If the value for nationalIdentifierType is not ‘LEIX’,
        // RegistrationAuthority MUST be populated.

        // A legal person must NOT have a value for countryOfIssue.
        "national_identification": {

            // Required: identifier issued by an appropriate issuing authority
            "national_identifier": "___",

            // Required: type of identifier specified.
            // Options include:
            //     0 : Alien registration number, assigned by a government agency.
            //     1 : Passport number, assigned by a passport authority.
            //     2 : Registration authority identifier.
            //     3 : Driver license number, assigned to a driver's license.
            //     4 : Foreign investment identity number, assigned to foreign
            //     investor.
            //     5 : Tax identification number, assigned by tax authority to
            //     entity.
            //     6 : Social security number, assigned by a social security
            //     agency.
            //     7 : Identity card number, assigned by a national authority.
            //     8 : Legal Entity Identifier, assigned in accordance with
            //     ISO 17442
            //     Note: The LEI is a 20-character, alpha-numeric code that
            //     enables clear and unique identification of legal entities
            //     participating in financial transactions.
            //     9 : Unspecified, a national identifier which may be known
            //     but which cannot otherwise be categorized or the category
            //     of which the sender is unable to determine.
            "national_identifier_type": ___,

            // Optional: Country of the issuing authority.
            "country_of_issue": "___",

            // Optional: A code specifying the registration authority.
            "registration_authority": "___"
        },

        // Optional: The country in which the legal person is registered.
        "country_of_registration": "___"
    },

    // TRIXO Questionnaire Details (REQUIRED)
    "trixo": {
        // Should be the name of the country or an ISO-3166-1 code.
        "primary_national_jurisdiction": "___",

        // Name of primary financial regulator or supervisory authority.
        "primary_regulator": "___",

        // Is the VASP permitted to send and/or receive transfers of virtual assets in
        // the jurisdictions in which it operates?
        // Options include:
        //     yes
        //     no
        //     partially
        "financial_transfers_permitted": "___",

        // Other jurisdictions in which the entity operates.
        "other_jurisdictions" : [
            {
                 "country" : "___",
                 "regulator_name": "___",
                 "license_number": "___"
            },
            {
                "country": "___",
                "regulator_name": "___",
                "license_number": "___"
            }
        ],

        // Does the VASP have a programme that sets minimum AML, CFT, KYC/CDD
        // and sanctions standards per the requirements of the jurisdiction(s)
        // regulatory regimes where it is licensed/approved/registered?
        // Options include:
        //     yes
        //     no
        "has_required_regulatory_program": "___",

        // Does the VASP conduct KYC/CDD before permitting its customers to
        // send/receive virtual asset transfers?
        // Options include:
        //     true
        //     false
        "conducts_customer_kyc": ___,

        // At what threshold does the VASP conduct KYC?
        // Floating point number, please note the currency
        "kyc_threshold": ___,
        "currency": "___",

        // Is the VASP required to comply with the application of Travel Rule
        // standards in the jurisdiction(s) where it is
        // licensed/approved/registered?
        // Options include:
        //     true
        //     false
        "must_comply_travel_rule": ___,

        // Applicable Travel Regulations the VASP must comply with.
        // Can list multiple regulations if relevant
        "applicable_regulations": [
            "___",
            "___"
        ],

        // What is the minimum threshold for travel rule compliance?
        // Floating point number
        "compliance_threshold": ___,

        // Is the VASP required by law to safeguard PII?
        // Options include:
        //     true
        //     false
        "must_safeguard_pii": ___,

        // Does the VASP secure and protect PII, including PII received
        // from other VASPs under the Travel Rule?
        // Options include:
        //     yes
        //     no
        "safeguards_pii": "___"
    }
}
```


NOTE: The default TRISA DS endpoint for the method is the TestNet instance (`api.vaspdirectory.net:443`)

This registration will result in an email being sent to all the technical contacts specified in the JSON file.  The emails will guide you through the remainder of the registration process.  Once you’ve completed the registration steps, TRISA TestNet administrators will receive your registration for review.

Once TestNet administrators have reviewed and approved the registration, you will receive a KYV certificate via email and your VASP will be publicly visible in the TestNet DS.


## Implementing the Trisa P2P Protocol


### Prerequisites

To begin setup, you’ll need the following:



*   KYV certificate (from TRISA DS registration)
*   The public key used for the CSR to obtain your certificate
*   The associated private key
*   The host name of the TRISA directory service
*   Ability to bind to the address:port that is associated with your VASP in the TRISA directory service.


### Integration Overview

Integrating the TRISA protocol involves both a client component and server component.

The client component will interface with a TRISA Directory Service (DS) instance to lookup other VASPs that integrate the TRISA messaging protocol.  The client component is utilized for outgoing transactions from your VASP to verify the receiving VASP is TRISA compliant.

The server component receives requests from other VASPs that integrate the TRISA protocol and provides responses to their requests.  The server component provides callbacks that must be implemented so that your VASP can return information satisfying the TRISA Network protocol.

Currently, a reference implementation of the TRISA Network protocol is available in Go.

[https://github.com/trisacrypto/testnet/blob/main/pkg/rvasp/trisa.go](https://github.com/trisacrypto/testnet/blob/main/pkg/rvasp/trisa.go)

Integrating VASPs must run their own implementation of the protocol.  If a language beside Go is required, client libraries may be generated from the protocol buffers that define the TRISA Network protocol.

Integrators are expected to integrate incoming transfer requests and key exchanges and may optionally also integrate outgoing transfer requests and key exchanges.

### Integration Notes

The TRISA Network protocol defines how data is transferred between participating VASPs.  The recommended format for data transferred for identifying information is the IVMS101 data format.  It is the responsibility of the implementing VASP to ensure the identifying data sent/received satisfies the FATF Travel Rule.

The result of a successful TRISA transaction results in a key and encrypted data that satisfies the FATF Travel Rule.  TRISA does not define how this data should be stored once obtained.  It is the responsibility of the implementing VASP to handle the secure storage of the resulting data for the transaction.


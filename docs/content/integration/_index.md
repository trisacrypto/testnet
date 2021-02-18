---
title: "VASP Integration"
date: 2020-02-17T21:00:00+08:00
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

After compiling the go protocol buffers per the documentation, you can run the following command to start the registration process:\

```
$ go run ./cmd/trisads register <json file>
```

The JSON file includes the registration information for your VASP in the TRIXO questionnaire format.  For a sample JSON file representing the TRIXO questionnaire:

```
{
    "trisa_endpoint": "api.alice.vaspbot.net:443",
    "website": "https://alice.vaspbot.net/",
    "business_category": 1,
    "vasp_category": 2,
    "established_on": "2020-11-10",
    "contacts": {
        "technical": {
            "name": "Benjamin Bengfort",
            "email": "benjamin@trisa.io"
        },
        "administrative": {
            "name": "Peter Hsu",
            "email": "peter@trisa.io"
        }
    },
    "entity": {
        "name": {
            "name_identifiers": [
                {
                    "legal_person_name": "AliceCoin, Inc.",
                    "legal_person_name_identifier_type": 0
                },
                {
                    "legal_person_name": "Alice VASP",
                    "legal_person_name_identifier_type": 1
                },
                {
                    "legal_person_name": "AliceCoin",
                    "legal_person_name_identifier_type": 2
                }
            ]
        },
        "geographic_addresses": [
            {
                "address_type": 1,
                "building_number": "23",
                "street_name": "Roosevelt Place",
                "town_name": "Boston",
                "country_sub_division": "MA",
                "post_code": "02151",
                "country": "US"
            }
        ],
        "national_identification": {
            "national_identifier": "5493004YBI24IF4TIP92",
            "national_identifier_type": 8,
            "country_of_issue": "US",
            "registration_authority": "RA000744"
        },
        "country_of_registration": "US"
    },
    "trixo": {
        "primary_national_jurisdiction": "US",
        "primary_regulator": "FinCEN",
        "financial_transfers_permitted": "yes",
        "has_required_regulatory_program": "yes",
        "conducts_customer_kyc": true,
        "kyc_threshold": 10000.00,
        "must_comply_travel_rule": true,
        "applicable_regulations": [
            "BSA Travel Rule – 31 CFR 103.33(g)",
            "FATF Travel Rule"
        ],
        "compliance_threshold": 10000.00,
        "must_safeguard_pii": false,
        "safeguards_pii": "yes"
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


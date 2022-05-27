#!/bin/bash

# rVASP common names
ALICE_NAME="api.alice.vaspbot.net"
BOB_NAME="api.bob.vaspbot.net"
EVIL_NAME="api.evil.vaspbot.net"

# rVASP UUIDs in GDS
VASP_PREFIX="vasps::"
ALICE_ID="alice0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0"
BOB_ID="bob0b0b0-b0b0-b0b0-b0b0-b0b0b0b0b0b0"
EVIL_ID="evile0e0-e0e0-e0e0-e0e0-e0e0e0e0e0e0"

# rVASP endpoints
ALICE_ENDPOINT="alice"
BOB_ENDPOINT="bob"
EVIL_ENDPOINT="evil"
TRISA_PORT="4435"

# GDS database DSN
GDS_DSN="leveldb:///fixtures/db"

# country code for local certificate generation
COUNTRY_CODE="US"

# Install required binaries: certs, gdsutil
go install github.com/trisacrypto/directory/cmd/certs@latest
go install github.com/trisacrypto/directory/cmd/gdsutil@latest

# Recreate the fixtures directory
if [ -e fixtures ]; then
    rm -rf fixtures
fi
mkdir -p fixtures

# Generate the rVASP certitifcates
mkdir -p fixtures/certs

certs init -c fixtures/certs/ca.gz
mkdir -p fixtures/certs/alice
certs issue -c fixtures/certs/ca.gz -o fixtures/certs/alice/cert.pem -n $ALICE_ENDPOINT -O localhost -C $COUNTRY_CODE
mkdir -p fixtures/certs/bob
certs issue -c fixtures/certs/ca.gz -o fixtures/certs/bob/cert.pem -n $BOB_ENDPOINT -O localhost -C $COUNTRY_CODE

certs init -c fixtures/certs/ca.evil.gz
mkdir -p fixtures/certs/evil
certs issue -c fixtures/certs/ca.evil.gz -o fixtures/certs/evil/cert.pem -n $EVIL_ENDPOINT -O localhost -C $COUNTRY_CODE

# Generate the rVASP fixtures from the template
mkdir -p fixtures/vasps
python scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/vasps/alice.json -n $ALICE_NAME -p $ALICE_ENDPOINT -i $ALICE_ID -e $ALICE_ENDPOINT:$TRISA_PORT
python scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/vasps/bob.json -n $BOB_NAME -p $BOB_ENDPOINT -i $BOB_ID -e $BOB_ENDPOINT:$TRISA_PORT
python scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/vasps/evil.json -n $EVIL_NAME -p $EVIL_ENDPOINT -i $EVIL_ID -e $EVIL_ENDPOINT:$TRISA_PORT

# Store the fixtures in the GDS database
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$ALICE_ID fixtures/vasps/alice.json
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$BOB_ID fixtures/vasps/bob.json
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$EVIL_ID fixtures/vasps/evil.json

# Confirm the keys in the database
echo ""
echo "Keys in the generated GDS database:"
gdsutil ldb:keys -d $GDS_DSN
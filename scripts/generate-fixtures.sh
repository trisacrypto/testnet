#!/bin/bash

# rVASP UUIDs in GDS
VASP_PREFIX="vasps::"
ALICE_ID="alice0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0"
BOB_ID="bob0b0b0-b0b0-b0b0-b0b0-b0b0b0b0b0b0"
EVIL_ID="evile0e0-e0e0-e0e0-e0e0-e0e0e0e0e0e0"
CHARLIE_ID="charlie0-c0c0-c0c0-c0c0-c0c0c0c0c0c0"

# rVASP common names - these should match the service names in docker compose
ALICE_NAME="alice"
BOB_NAME="bob"
EVIL_NAME="evil"
CHARLIE_NAME="charlie"
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
certs issue -c fixtures/certs/ca.gz -o fixtures/certs/alice/cert.pem -n $ALICE_NAME -O localhost -C $COUNTRY_CODE
mkdir -p fixtures/certs/bob
certs issue -c fixtures/certs/ca.gz -o fixtures/certs/bob/cert.pem -n $BOB_NAME -O localhost -C $COUNTRY_CODE
mkdir -p fixtures/certs/charlie
certs issue -c fixtures/certs/ca.gz -o fixtures/certs/charlie/cert.pem -n $CHARLIE_NAME -O localhost -C $COUNTRY_CODE

certs init -c fixtures/certs/ca.evil.gz
mkdir -p fixtures/certs/evil
certs issue -c fixtures/certs/ca.evil.gz -o fixtures/certs/evil/cert.pem -n $EVIL_NAME -O localhost -C $COUNTRY_CODE

# Generate the GDS fixtures from the template
mkdir -p fixtures/gds
python3 scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/gds/alice.json -n $ALICE_NAME -i $ALICE_ID -c fixtures/certs/alice/cert.pem -e $ALICE_NAME:$TRISA_PORT
python3 scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/gds/bob.json -n $BOB_NAME -i $BOB_ID -c fixtures/certs/bob/cert.pem -e $BOB_NAME:$TRISA_PORT
python3 scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/gds/evil.json -n $EVIL_NAME -i $EVIL_ID -c fixtures/certs/evil/cert.pem -e $EVIL_NAME:$TRISA_PORT
python3 scripts/fixtures/gds-fixture.py -t scripts/fixtures/template.json -o fixtures/gds/charlie.json -n $CHARLIE_NAME -i $CHARLIE_ID -c fixtures/certs/charlie/cert.pem -e $CHARLIE_NAME:$TRISA_PORT

# Migrate the rVASP fixtures to the fixtures directory
mkdir -p fixtures/rvasps
python3 scripts/fixtures/rvasp-fixtures.py -f pkg/rvasp/fixtures -o fixtures/rvasps -n $ALICE_NAME,$BOB_NAME,$EVIL_NAME,$CHARLIE_NAME

# Store the rVASP records in the GDS database
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$ALICE_ID fixtures/gds/alice.json
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$BOB_ID fixtures/gds/bob.json
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$EVIL_ID fixtures/gds/evil.json
gdsutil ldb:put -d $GDS_DSN $VASP_PREFIX$CHARLIE_ID fixtures/gds/charlie.json

# Confirm the keys in the database
echo ""
echo "Keys in the generated GDS database:"
gdsutil ldb:keys -d $GDS_DSN
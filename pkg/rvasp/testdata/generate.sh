#!/bin/bash

# Remove old key if it exists
keyFiles=("cert.pem" "ca.gz")
for keyfile in ${keyFiles}[@]}; do
    if [ -f $keyfile ]; then
        rm $keyfile
    fi
done

# Create CA
openssl req -x509 -newkey rsa:4096 -sha256 -days 10950 \
    -nodes -keyout ca.key -out ca.crt \
    -subj "/C=US/ST=California/L=Menlo Park/O=TRISA/OU=TestNet/CN=trisatest.dev" \
    -addext "subjectAltName=DNS:trisatest.dev,DNS:*.trisatest.dev"

# Create certificate requests for alice and bob
openssl req -new -newkey rsa:4096 \
    -nodes -keyout alice.key.pem -out alice.csr \
    -subj "/C=US/ST=New York/L=New York/O=Alice VASP/OU=Testing/CN=alice" \
    -addext "subjectAltName=DNS:alice.vaspbot.net,DNS:*.alice.vaspbot.net,DNS:bufnet,DNS:alice"

# Create signed certificates with CA
openssl x509 -req -days 10950 \
    -CA ca.crt -CAkey ca.key \
    -in alice.csr -out cert.pem \
    -copy_extensions copyall

# Combine files into a single certificate chain
cat ca.crt >> cert.pem
cat alice.key.pem >> cert.pem
mv ca.crt ca
gzip ca

# Cleanup
rm alice.csr alice.key.pem
rm ca.key
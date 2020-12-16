# TRISA Directory Service

**Implements a simple gRPC directory service for TRISA.**

This is a prototype implementation of a gRPC directory service that can act as a standalone server for VASP lookup queries. This is not intended to be used for production, but rather as a proof-of-concept (PoC) for directory service registration, lookups, and searches.

## Generate Protocol Buffers

To regenerate the Go and Javascript code from the protocol buffers:

```
$ go generate ./...
```

The go generate directives are stored in `pb/pb.go`. The directives create grpc Go in the `pb` package as well as grpc-web in the `web/src/pb` directory.

## Quick Start

The simplest way to get started is to use the `docker-compose.yml` file to get the following services running locally:

- **trsisads**: the TRISA directory service that implements the grpc protocol
- **proxy**: an envoy proxy that translates HTTP 1.1 requests into HTTP 2.0 requests
- **dsui**: UI that implements grpc-web to connect to the directory server via the proxy

Run the services as follows:

```
$ docker-compose up
```

Then connect to the UI on https://localhost:8000/

### Development

For development purposes you'll want to run and reload the servers individually. To run the directory service:

```
$ go run ./cmd/trisads serve
```

Note that you'll likely want to have the following environment variables configured:

- `$SECTIGO_USERNAME`, `$SECTIGO_PASSWORD`: to access the Sectigo API
- `$SENDGRID_API_KEY`: sending verification emails and certificates

To run the development web UI server:

```
$ cd web
$ npx serve
```

Finally, to run the proxy, use the docker image, building if necessary:

```
$ docker run -n grpc-proxy trisacrypto/proxy:develop
```

## Sectigo API

The directory service issues certificates using Sectigo, please refer to the [API Documentation](https://support.sectigo.com/Com_KnowledgeDetailPage?Id=kA01N000000bvCJ) for more details on the endpoints and supported interactions. The `sectigo` package provides a simple client interface for interacting with the API; most of this code is handled by the server, but there is also a CLI interface that demonstrates usage.

To install the CLI server:

```
$ go install ./cmd/sectigo
```

This should add the `sectigo` command to your `$PATH`.

### Authentication

The first step is authentication, you should set your username and password in the `$SECTIGO_USERNAME` and `$SECTIGO_PASSWORD` environment variables (alternatively you can pass them as parameters on the command line). To verify your authentication status you can use:

```
$ sectigo auth
```

The API authenticates by username and password then returns acess and refresh tokens which are stored in a local cache file. To see where your cache is stored:

```
$ sectigo auth --cache
```

If you'd like to check your credentials state, e.g. if the access tokens are valid, refreshable, or expired, use:

```
$ sectigo auth --debug
```

### Authorities and Profiles

To begin to interact with certificates you need to list the authorities and profiles that your user account has access to.

```
$ sectigo authorities
[
  {
    "id": 1,
    "ecosystemId": 100,
    "signerCertificateId": 0,
    "ecosystemName": "TRISA",
    "balance": 10,
    "enabled": true,
    "profileId": 42,
    "profileName": "TRISA Profile"
  }
]
```

The authority displays the methods and profiles that certificates are created under. Here the `profileId` field is very important for use in subsequent calls. You can also view how many licenses have been ordered/issued across all authorities as follows:

```
$ sectigo licenses
{
  "ordered": 2,
  "issued": 2
}
```

To get detail information for a profile, use the profile ID with the following command:

```
$ sectigo profiles -i 42
```

This will return the raw profile configuration. Before creating certificates with the authority, you'll need to know the required profile parameters:

```
$ sectigo profile -i 42 --params
```

### Creating Certificates

You can request a certificate to be created with the `commonName` and `pkcs12Password` params as follows (note for profiles that require other params, you'll have to use the code base directly and implement your own method):

```
$ sectigo create -a 42 -d example.com -p secrtpasswrd -b "example.com certs"
{
  "batchId": 24,
  "orderNumber": 1024,
  "creationDate": "2020-12-10T16:35:32.805+0000",
  "profile": "TRISA Profile",
  "size": 1,
  "status": "CREATED",
  "active": false,
  "batchName": "example.com certs",
  "rejectReason": "",
  "generatorParametersValues": null,
  "userId": 10,
  "downloadable": true,
  "rejectable": true
}
```

The `-a` flag specifies the authority, but should be a profile id. The domain must be a valid domain. If you don't specify a password, one is generated for you and printed on the CLI before exit. The `-b` flag gives a human readable name for the batch creation. The return data shows detail about the batch certificate job that was created; you can fetch the data to keep checking on the status as follows:

```
$ sectigo batches -i 24
```

You can also get processing information for the batch:

```
$ sectigo batches -i 24 --status
```

Once the batch is created, it's time to download the certificates in a ZIP file:

```
$ sectigo download -i 24 -o certs/
```

This will download the batch file (usually batchId.zip, 24.zip in this case) to the `certs/` directory. Unzip the certs then decrypt the .pem file as follows:

```
$ unzip certs/24.zip
$ openssl pkcs12 -in certs/example.com.p12 -out certs/example.com.pem -nodes
```

For more on working with the PKCS12 file, see [Export Certificates and Private Key from a PKCS#12 File with OpenSSL](https://www.ssl.com/how-to/export-certificates-private-key-from-pkcs12-file-with-openssl/).

### Managing Certificates

You can search for a certificate by name or serial number, but mostly commonly you search by the domain or common name to get the serial number:

```
$ sectigo find -n example.com
```

Once you've obtained the serial number you can revoke the certificate as follows:

```
$ sectigo revoke -p 42 -r "cessation of operation" -s 12345
```

This command expects the profile id that issued the certificate with the `-p` flag, an [RFC 5280 reason code](https://tools.ietf.org/html/rfc5280#section-5.3.1) passed via the `-r` flag (unspecified by default), and the serial number of the certificate using the `-s` flag. If this command doesn't error, then the certificate has been successfully revoked.

The RFC 5280 reasons are:

- "unspecified"
- "keycompromise"
- "ca compromise"
- "affiliation changed"
- "superseded"
- "cessation of operation"
- "certificate hold"
- "remove from crl"
- "privilege withdrawn"
- "aa compromise"

Note that the reason is whitespace and case insensitive.

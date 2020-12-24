# TRISA TestNet

**An integration and test platform for the [TRISA Travel Rule Implementation](https://trisa.io).**

The TRISA test net is comprised of the following services:

- [TRISA Directory Service](https://api.vaspdirectory.net) - a grpc service for registering TRISA participants, issuing certificates, and establishing trust.
- [TRISA Directory UI](https://vaspdirectory.net) - a user interface to explore the TRISA directory service and interact with it manually.
- [Envoy gRPC Proxy](https://proxy.vaspdirectory.net) - facilitates grpc-web interactions with the TRISA directory service.
- [Alice rVASP](https://alice.vaspbot.net) - a "robot VASP" that demonstrates TRISA transactions and acts as an integration backstop to initiate and deliver transactions to.
- [Bob rVASP](https://bob.vaspbot.net) - a secondary "robot VASP" that demos interactions with Alice and can develop against constrained resources.
- [Evil rVASP](https://evil.vaspbot.net) - a "robot VASP" that is not registered with TRISA and is used to ensure correct error handling when the protocol is used incorrectly.

## Monorepo Organization

This repository consists of a monorepo that is designed to facilitate all TRISA test net operations. For now it is the reference implementation of TRISA until we can integrate it into the [trisacrypto/trisa](https://github.com/trisacrypto/trisa) repository. The organzation of the repository is as follows:

- `cmd`: binary executables compiled with go
- `containers`: Dockerfiles for various containers deployed to the test net
- `docs`: documentation built with hugo
- `fixtures`: initial or example data used to bootstrap services
- `lib`: library modules in other languages (e.g. Python)
- `manifests`: kubernetes manifests for our GKE cluster
- `pkg`: Go code and implementations for various services
- `proto`: Protocol Buffer definitions for the services
- `web`: front-end web applications, either pure HTML or npm based

## Generate Protocol Buffers

To regenerate the Go and Javascript code from the protocol buffers:

```
$ go generate ./...
```

The go generate directives are stored in `pb/pb.go`. The directives create grpc Go in the `pb` package as well as grpc-web in the `web/src/pb` directory.

## Directory Service

This is a prototype implementation of a gRPC directory service that can act as a standalone server for VASP lookup queries. This is not intended to be used for production, but rather as a proof-of-concept (PoC) for directory service registration, lookups, and searches.

Thre directory service is composed of three component services:

- **trsisads**: the TRISA directory service that implements the grpc protocol
- **proxy**: an envoy proxy that translates HTTP 1.1 requests into HTTP 2.0 requests
- **dsweb**: UI that implements grpc-web to connect to the directory server via the proxy

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
$ cd web/trisads
$ npx serve
```

Finally, to run the proxy, use the docker image, building if necessary:

```
$ docker run -n grpc-proxy trisa/grpc-proxy:develop
```
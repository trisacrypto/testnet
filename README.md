# TRISA TestNet

**An integration and test platform for the [TRISA Travel Rule Implementation](https://trisa.io).**

The TRISA test net is comprised of the following:

- [TRISA Directory Service](https://trisatest.net) - a user interface to explore the TRISA Global Directory Service and register to become a TRISA member
- [TestNet Demo](https://vaspbot.net) - a demo site to show TRISA interactions between “robot” VASPs that run in the TestNet

For more details, see the [TRISA Documentation](https://trisatest.net/), or check out the [TRISA codebase](https://github.com/trisacrypto/trisa).


## Monorepo Organization

This repository consists of a monorepo that is designed to facilitate all TRISA test net operations. The organization of the repository is as follows:

- `cmd`: binary executables compiled with go
- `containers`: Dockerfiles for various containers deployed to the test net
- `fixtures`: initial or example data used to bootstrap services
- `lib`: library modules in other languages (e.g. Python)
- `manifests`: kubernetes manifests for our GKE cluster
- `pkg`: Go code and implementations for various services
- `proto`: Protocol Buffer definitions for the services
- `scripts`: Shell, bash, and Python scripts used for local testing
- `web`: front-end web applications, either pure HTML or npm based

## Generate Protocol Buffers

To regenerate the Go and Javascript code from the protocol buffers:

```
$ go generate ./...
```

The go generate directives are stored in `pb/pb.go`. The directives create grpc Go in the `pb` package as well as grpc-web in the `web/src/pb` directory.
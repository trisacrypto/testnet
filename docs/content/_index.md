---
title: Home
draft: false
---

# TRISA Test Net

The TRISA test net is comprised of the following services:

- [TRISA Directory Service](https://api.vaspdirectory.net) - a grpc service for registering TRISA participants, issuing certificates, and establishing trust.
- [TRISA Directory UI](https://vaspdirectory.net) - a user interface to explore the TRISA directory service and interact with it manually.
- [Envoy gRPC Proxy](https://proxy.vaspdirectory.net) - facilitates grpc-web interactions with the TRISA directory service.
- [Alice rVASP](https://alice.vaspbot.net) - a "robot VASP" that demonstrates TRISA transactions and acts as an integration backstop to initiate and deliver transactions to.
- [Bob rVASP](https://bob.vaspbot.net) - a secondary "robot VASP" that demos interactions with Alice and can develop against constrained resources.
- [Evil rVASP](https://evil.vaspbot.net) - a "robot VASP" that is not registered with TRISA and is used to ensure correct error handling when the protocol is used incorrectly.

More documentation coming soon!
---
title: Home
date: 2020-12-24T07:58:37-05:00
lastmod: 2020-12-24T07:58:37-05:00
description: "Directory Service interactions with the Sectigo CA API"
weight: 10
---

# TRISA TestNet

The goal of the Travel Rule Information Sharing Architecture (TRISA) is to enable
compliance with the FATF and FinCEN Travel Rules for cryptocurrency transaction
identity information without modifying the core blockchain protocols, and without
incurring increased transaction costs or modifying virtual currency peer-to-peer
transaction flows.

{{% button status="notice" icon="fas fa-download" url="https://trisa.io/trisa-whitepaper/" %}}Read the TRISA White Paper v8{{% /button %}}

## About the TestNet

The TRISA TestNet has been established to provide a demonstration of the TRISA peer-to-peer protocol, host "robot VASP" services to facilitate TRISA integration, and is the location of the primary TRISA Directory Service that facilitates public key exchange and peer discovery.

{{% figure src="/img/testnet_architecture.png" %}}

The TRISA test net is comprised of the following services:

- [TRISA Directory Service](https://api.vaspdirectory.net) - a grpc service for registering TRISA participants, issuing certificates, and establishing trust.
- [TRISA Directory UI](https://vaspdirectory.net) - a user interface to explore the TRISA directory service and interact with it manually.
- [Envoy gRPC Proxy](https://proxy.vaspdirectory.net) - facilitates grpc-web interactions with the TRISA directory service.
- [Alice rVASP](https://alice.vaspbot.net) - a "robot VASP" that demonstrates TRISA transactions and acts as an integration backstop to initiate and deliver transactions to.
- [Bob rVASP](https://bob.vaspbot.net) - a secondary "robot VASP" that demos interactions with Alice and can develop against constrained resources.
- [Evil rVASP](https://evil.vaspbot.net) - a "robot VASP" that is not registered with TRISA and is used to ensure correct error handling when the protocol is used incorrectly.

More documentation coming soon!
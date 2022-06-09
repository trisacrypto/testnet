# rVASP

**Robot VASP for TRISA demonstration and integration**

This is a simple gRPC server that implements a UI workflow and messaging framework to demonstrate sending and receiving transactions using the TRISA InterVASP protocol. The server was built to support a demonstration UI that requires a streaming interface so that a live message flow is achieved. However, the communication protocol between rVASPs also demonstrates how implementers might go about writing InterVASP services in their own codebases.

## Generating Protocol Buffers

To regenerate the Go and Python code from the protocol buffers:

```
$ go generate ./...
```

This will generate the Go code in `pkg/rvasp/pb/v1` and the Python code in `lib/python/rvaspy/rvaspy`. Alternatively you can manually generate the code to specify different directories using the following commands:

```
$ protoc -I . --go_out=plugins=grpc:. api.proto
$ python -m grpc_tools.protoc -I . --python_out=./rvaspy --grpc_python_out=./rvaspy api.proto
```

Note: After generating the Python protocol buffers you must edit the import line in `lib/python/rvaspy/rvaspy/api_pb2_grpc.py` to the following in order to correctly use the library.

```
import rvaspy.api_pb2 as api__pb2
```

## Quick Start

To get started using the rVASP, you can run the local server as follows:

```
$ go run ./cmd/rvasp serve
```

Alternatively, multiple rVASPs can be run concurrently using docker compose.

```
$ ./scripts/generate-fixtures.sh
$ ./containers/build.sh
$ docker compose -f ./containters/docker-compose.yml up
```

The server should now be listening for TRISADemo RPC messages. To send messages using the python API, make sure you can import the modules from `rvaspy` - the simplest way to do this is to install the package in editable mode as follows:

```
$ pip install -e ./rvaspy/
```

This will use the `setup.py` file in the `rvaspy` directory to install the package to your `$PYTHON_PATH`. Because it is in editable mode, any time you regenerate the protocol buffers or pull the repository, the module should be updated on the next time you import. An example script for using the package is as follows:

```python
import rvaspy

api = rvaspy.connect("localhost:6434")

cmds = [
    api.account_request("robert@bobvasp.co.uk"),
    api.transfer_request("robert@bobvasp.co.uk", "mary@alicevasp.us", 42.99)
]

for msg in api.stub.LiveUpdates(iter(cmds)):
    print(msg)
```

Note that the `RVASP` api client is not fully implemented yet.

## Configuration

The identity and operation of the rVASPs is defined in the `pkg/rvasp/fixtures` directory. This must contain two files, `vasps.json` and `wallets.json`.

### vasps.json

This defines how rVASPs are addressed and their ivms101 identity information. It consists of a list of JSON objects with the fields:

`common_name`: Common name of the rVASP
`legal_person`: ivms101 identity information of the rVASP

### wallets.json

This defines the wallets and accounts in the rVASP database. It consists of a list of ordered entries for each wallet:

1. The wallet address
2. The email address of the associated account
3. The index of the VASP in `vasps.json`, starting at 1
4. The originator policy for outgoing transfers
5. The beneficiary policy for incoming transfers
6. The ivms101 information for the associated account

### Wallet Policies

The rVASPs are designed to support different configured transfer policies without having to rebuild them. This is implemented by associating wallets with policies. The supported policies are defined below:

#### Originator (Outgoing) Policies

`send_partial`: Send a transfer request to the beneficiary containing partial beneficiary identity information.

`send_full`: Send a transfer request to the beneficiary containing the full beneficiary identity information.

`send_error`: Send a TRISA error to the beneficiary.

#### Beneificiary (Incoming) Policies

`sync_repair`: Complete the beneficiary identity information in the received payload and return the payload to the originator.

`sync_require`: Send a synchronous response to the originator if the full beneficiary identity information exists in the payload, otherwise send a TRISA rejection error.

`async_repair`: Send a pending response to the originator with ReplyNotBefore and ReplyNotAfter timestamps. After a period of time within that time range, initiate a transfer to the originator containing the full beneficiary identity information to complete the transaction.

`async_reject`: Send a pending response to the originator with ReplyNotBefore and ReplyNotAfter timestamps. After a period of time within that time range, send a TRISA rejection error to the originator.
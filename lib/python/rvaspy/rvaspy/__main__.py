import argparse

from rvaspy import connect
from rvaspy.client import HOST, CLIENT
from rvaspy.version import get_version


def main(args):
    api = connect(host=args.addr, name=args.client)

    cmds = [
        api.account_request("alice@alicevasp.us"),
        api.transfer_request("alice@alicevasp.us", "robert@bobvasp.co.uk", 42.99)
    ]

    for msg in api.stub.LiveUpdates(iter(cmds)):
        print(msg)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="simple rvasp client")
    parser.add_argument("-v", "--version", action="version", version=f"rvaspy v{get_version()}")
    parser.add_argument(
        "-c", "--client", default=CLIENT, type=str, metavar="NAME",
        help="name of the client connecting to the server",
    )
    parser.add_argument(
        "-a", "--addr", default=HOST, type=str, metavar="HOST:PORT",
        help="address to connect to the rvasp server with"
    )

    args = parser.parse_args()
    main(args)
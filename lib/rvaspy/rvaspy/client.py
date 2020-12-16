#!/usr/bin/env python3

import grpc

from .api_pb2 import *  # noqa
from .api_pb2_grpc import *  # noqa


CLIENT = "rvaspy"
HOST = "localhost:4434"
RFC3339 = "%Y-%m-%dT%H:%M:%S.%fZ"


class RVASP(object):
    """
    An API wrapper for accessing the TRISA Demo rVASP Server.
    """

    # increment message number at the class level
    _msgseq = 0

    def __init__(self, name=CLIENT, host=HOST):
        self.channel = grpc.insecure_channel(host)
        self.stub = TRISADemoStub(self.channel)
        self.name = name

    def _wrap_command(self, rpc, request):
        """
        Helper function to wrap an account status or transfer request into a streaming
        command to actively listen
        """
        self._msgseq += 1
        kwargs = {
            "type": rpc,
            "id": self._msgseq,
            "client": self.name,
        }

        if rpc == ACCOUNT:
            kwargs["account"] = request
        elif rpc == TRANSFER:
            kwargs["transfer"] = request

        return Command(**kwargs)

    def account_request(self, account, no_transactions=False):
        """
        Creates an account request command for sending via streaming.

        Parameters
        ----------
        account : str
            email address of the account to get information for

        no_transactions : bool
            does not return transactions list if true
        """
        req = AccountRequest(account=account, no_transactions=no_transactions)
        return self._wrap_command(ACCOUNT, req)

    def transfer_request(self, account, beneficiary, amount):
        """
        Creates a transfer request command for sending via streaming.

        Parameters
        ----------
        account : str
            email address of account to debit

        beneficiary : str
            email address or wallet id to look up beneficiary with

        amount : float
            amount to transfer to the beneficiary
        """
        req = TransferRequest(account=account, beneficiary=beneficiary, amount=amount)
        return self._wrap_command(TRANSFER, req)

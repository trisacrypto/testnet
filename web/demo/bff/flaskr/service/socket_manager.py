# Tracks sockets and the context that they're subscribed to
# context is tied to the end user's session.
# User could be viewing multiple VASPs UIs in the same context.
# This allows one user to perform actions across multiple vasps and each UI session receive appropriate data
# for that context.
import json

from flask import request
from flask_socketio import SocketIO, emit, join_room, leave_room, rooms
from google.protobuf.json_format import MessageToJson

from flaskr.db import query_vasp
from flaskr.models.transaction import Transaction
from flaskr.models.transaction_request import TransactionRequest
from flaskr.models.vasp_context import VaspContext
from flaskr.models.vasp_log_message import VaspLogMessage
from flaskr.rvaspy.rvaspy import RVASP
from flaskr.simulator.transaction_handler import TransactionHandler


class SocketManager:

    def __init__(self, socketio: SocketIO, transaction_handler: TransactionHandler):

        self.api = RVASP(name="")
        self.vasp_context = VaspContext("", "")

        @socketio.on('transaction_request')
        def handle_transaction_request(message):
            transaction_request = TransactionRequest.from_json(message)

            print('Received transaction request ' + message)

            # transfer_request = self.api.transfer_request(transaction_request.originator_wallet_id,
            #                                              transaction_request.beneficiary_wallet_id,
            #                                              transaction_request.amount)
            transfer_request = self.api.transfer_request("alice@alicevasp.us",
                                                         "robert@bobvasp.co.uk",
                                                         transaction_request.amount)

            for msg in self.api.stub.LiveUpdates(iter([transfer_request])):
                self.handle_message(msg)

        @socketio.on('vasp_context')
        def handle_vasp_context(message):
            self.clear_context_for_session(request.sid)

            context = VaspContext.from_json(message)
            self.vasp_context = context
            # TODO: handle context errors
            self.set_context_for_session(context, request.sid)

            # Create GRPC connection to rvasp
            vasp = query_vasp(context.vasp_id)[0]
            self.api = RVASP(name="client", host=vasp['websocket_address'])

        @socketio.on('connect')
        def connected():
            # Don't really do anything since no room assigned yet
            pass

        @socketio.on('disconnect')
        def disconnect():
            # Not really needed with rooms?
            self.clear_context_for_session(request.sid)

    def handle_message(self, msg):
        print(msg)
        if msg.type:
            print('received msg type')
            print(msg.type)
            if msg.type == 1:
                print('received msg type TRANSFER')
                self.handle_transaction_message(msg)
        else:
            self.handle_log_message(msg)

    def handle_log_message(self, msg):
        self.broadcast_to_context(
            self.vasp_context,
            'vasp_log_message',
            VaspLogMessage(self.vasp_context.vasp_id, msg.timestamp, '',
                           msg.update, self.map_category_to_color(msg.category))
        )

    def map_category_to_color(self, category):
        if category == 0:  # 'LEDGER'
            return '888888'
        elif category == 1:  # 'TRISADS':
            return '99ccff'
        elif category == 2:  # 'TRISAP2P':
            return '0080ff'
        elif category == 3:  # 'BLOCKCHAIN':
            return '00cc66'
        elif category == 4:  # 'ERROR':
            return 'cc0000'
        else:
            return 'ffffff'

    def handle_transaction_message(self, msg):
        self.broadcast_to_context(
            self.vasp_context,
            'transaction',
            Transaction(
                msg.transfer.transaction.timestamp,
                '23hlkjad824',  # TODO: need transaction ID?
                msg.transfer.transaction.originator.wallet_address,
                'BOB-GUID',
                'BobVASP',  # TODO: need basic VASP info?
                msg.transfer.transaction.beneficiary.wallet_address,
                'ALICE-GUID',
                'AliceVASP',
                MessageToJson(msg.transfer)
            ))

    # Sets a new context for a connected socket
    def set_context_for_session(self, context: VaspContext, session_id: str):
        join_room(context.get_room_identifier(), session_id)

    # Removes any context for a connected socket
    def clear_context_for_session(self, session_id: str):
        for room in rooms(session_id):
            leave_room(room, session_id)

    # Broadcast messages to all sockets listening to a context
    def broadcast_to_context(self, context: VaspContext, event, message):
        emit(event, json.dumps(message.__dict__), room=context.get_room_identifier())

# Tracks sockets and the context that they're subscribed to
# context is tied to the end user's session.
# User could be viewing multiple VASPs UIs in the same context.
# This allows one user to perform actions across multiple vasps and each UI session receive appropriate data
# for that context.
import json
import time

from flask import request
from flask_socketio import SocketIO, emit, join_room, leave_room, rooms
from google.protobuf.json_format import MessageToJson
from rvaspy import RVASP

from flaskr.db import query_vasp
from flaskr.models.transaction import Transaction
from flaskr.models.transaction_request import TransactionRequest
from flaskr.models.vasp_context import VaspContext
from flaskr.models.vasp_log_message import VaspLogMessage
from flaskr.simulator.transaction_handler import TransactionHandler


def blocking_listener(api, timeout=36, iters=100):
    yield api.norpc_request()
    for _ in range(iters):
        time.sleep(timeout)


class SocketManager:

    def __init__(self, socketio: SocketIO, transaction_handler: TransactionHandler):

        self.originator_api = None
        self.beneficiary_api = None
        self.originator_vasp_context = None
        self.beneficiary_vasp_context = None

        @socketio.on('transaction_request')
        def handle_transaction_request(message):
            transaction_request = TransactionRequest.from_json(message)

            print('Received transaction request ' + message)

            transfer_request = self.originator_api.transfer_request(transaction_request.originator_wallet_id,
                                                                    transaction_request.beneficiary_wallet_id,
                                                                    transaction_request.amount,
                                                                    transaction_request.originator_vasp_id,
                                                                    transaction_request.beneficiary_vasp_id)

            print('Sending transfer to vasp ' + self.originator_api.name +
                  ' request originator:' + transaction_request.originator_wallet_id +
                  ' beneficiary:' + transaction_request.beneficiary_wallet_id +
                  ' originating vasp:' + transaction_request.originator_vasp_id +
                  ' beneficiary vasp:' + transaction_request.beneficiary_vasp_id)

            # subscribe to all updates
            for msg in self.originator_api.stub.LiveUpdates(iter([transfer_request])):
                self.handle_message(self.originator_vasp_context, msg)
            for msg in self.beneficiary_api.stub.LiveUpdates(iter([self.beneficiary_api.norpc_request()])):
                self.handle_message(self.beneficiary_vasp_context, msg)

        @socketio.on('vasp_context')
        def handle_vasp_context(message):
            context = VaspContext.from_json(message)
            self.clear_context_for_session(request.sid)
            self.set_context_for_session(context, request.sid)
            vasp = query_vasp(context.vasp_id)[0]

            if context.originator:
                self.originator_vasp_context = context
                print("Received originator vasp context " + context.vasp_id + " creating client to " +
                      vasp['websocket_address'])
                self.originator_api = RVASP(name=context.vasp_id, host=vasp['websocket_address'])

                # subscribe to all updates
                for msg in self.originator_api.stub.LiveUpdates(iter([self.originator_api.norpc_request()])):
                    self.handle_message(context, msg)
            else:
                self.beneficiary_vasp_context = context
                print("Received Beneficiary vasp context " + context.vasp_id + " creating client to " +
                      vasp['websocket_address'])
                self.beneficiary_api = RVASP(name=context.vasp_id, host=vasp['websocket_address'])

                # subscribe to all updates
                for msg in self.beneficiary_api.stub.LiveUpdates(blocking_listener(self.beneficiary_api)):
                    self.handle_message(context, msg)

        @socketio.on('connect')
        def connected():
            # Don't really do anything since no room assigned yet
            pass

        @socketio.on('disconnect')
        def disconnect():
            # Not really needed with rooms?
            self.clear_context_for_session(request.sid)

    def handle_message(self, vasp_context: VaspContext, msg):
        if msg.type:
            if msg.type == 1:
                self.handle_transaction_message(vasp_context, msg)
        else:
            self.handle_log_message(vasp_context, msg)

    def handle_log_message(self, vasp_context: VaspContext, msg):
        self.broadcast_to_context(
            vasp_context,
            'vasp_log_message',
            VaspLogMessage(vasp_context.vasp_id, msg.timestamp, '',
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

    def handle_transaction_message(self, vasp_context: VaspContext, msg):
        self.broadcast_to_context(
            vasp_context,
            'transaction',
            Transaction(
                msg.transfer.transaction.timestamp,
                '23hlkjad824',  # TODO: need transaction ID - Issue #1
                msg.transfer.transaction.originator.wallet_address,
                'api.bob.vaspbot.net',
                msg.transfer.transaction.originator.provider,
                msg.transfer.transaction.beneficiary.wallet_address,
                'api.alice.vaspbot.net',
                msg.transfer.transaction.beneficiary.provider,
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

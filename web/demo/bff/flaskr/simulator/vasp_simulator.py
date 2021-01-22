from flaskr.models.transaction_request import TransactionRequest
from flaskr.models.vasp_details import VaspDetails
from flaskr.simulator.transaction_handler import TransactionHandler


class VaspSimulator(TransactionHandler):
    vasp_details: VaspDetails

    def __init__(self, vasp_details: VaspDetails):
        self.vasp_details = vasp_details
        pass
        # TODO: initialize client grpc streaming protocols

    def handle_transaction_request(self, transaction_request: TransactionRequest):
        # TODO: start transaction

        pass

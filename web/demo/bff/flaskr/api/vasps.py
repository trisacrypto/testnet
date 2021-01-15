import jsons
from flask_restful import Resource

from flaskr.db import get_db, query_all_vasps, query_wallets
from flaskr.models.trisa_ds_entry import TrisaDsEntry
from flaskr.models.user_wallet import UserWallet
from flaskr.models.vasp_details import VaspDetails


class Vasps(Resource):
    def get(self):

        vasps = []

        for vasp in query_all_vasps():
            vasp_id = vasp['vasp_id']

            wallets = []

            for wallet in query_wallets(vasp_id):
                wallets.append(
                    UserWallet(
                        wallet['wallet_id'],
                        wallet['wallet_address']
                    )
                )

            trisa_ds_entry = None

            if vasp['trisa_ds_id']:
                trisa_ds_entry = TrisaDsEntry(
                    vasp['trisa_ds_id'],
                    vasp['trisa_ds_name'],
                    vasp['trisa_protocol_host']
                )

            vasps.append(
                VaspDetails(
                    vasp['vasp_id'],
                    vasp['display_name'],
                    vasp['description'],
                    trisa_ds_entry,
                    vasp['private_key'],
                    vasp['public_key'],
                    wallets
                )
            )

        return jsons.dump({'data': vasps})

from dataclasses import dataclass

from dataclasses_json import dataclass_json

from flaskr.models.vasp_details import VaspDetails


@dataclass_json
@dataclass
class Transaction:
    timestamp: str
    transaction_id: str
    originating_wallet: str
    originating_vasp_id: str
    originating_vasp_display_name: str
    beneficiary_wallet: str
    beneficiary_vasp_id: str
    beneficiary_vasp_display_name: str
    ivms101Data: str

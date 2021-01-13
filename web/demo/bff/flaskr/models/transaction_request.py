from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class TransactionRequest:
    context_id: str
    originator_vasp_id: str
    originator_wallet_id: str
    beneficiary_vasp_id: str
    beneficiary_wallet_id: str
    crypto_type: str
    amount: float

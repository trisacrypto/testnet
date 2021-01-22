from dataclasses import dataclass
from typing import List, Optional

from dataclasses_json import dataclass_json

from flaskr.models.trisa_ds_entry import TrisaDsEntry
from flaskr.models.user_wallet import UserWallet


@dataclass_json
@dataclass
class VaspDetails:
    vasp_id: str
    display_name: str
    description: str
    trisa_ds_entry: Optional[TrisaDsEntry]
    private_key: str
    public_key: str
    user_wallets: List[UserWallet]

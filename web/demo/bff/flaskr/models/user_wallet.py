from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class UserWallet:
    user_wallet_id: str
    wallet_address: str

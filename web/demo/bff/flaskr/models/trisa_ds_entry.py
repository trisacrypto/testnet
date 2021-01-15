from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class TrisaDsEntry:
    trisa_ds_id: str
    display_name: str
    trisa_protocol_host: str

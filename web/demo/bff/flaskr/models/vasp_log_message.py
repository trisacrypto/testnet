from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class VaspLogMessage:
    vasp_id: str
    timestamp: str
    message: str
    message_unencrypted: str
    color_code: str

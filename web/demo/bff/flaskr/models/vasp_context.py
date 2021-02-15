from dataclasses import dataclass

from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class VaspContext:
    context_id: str
    vasp_id: str
    originator: bool

    def get_room_identifier(self):
        return self.context_id + ':' + self.vasp_id

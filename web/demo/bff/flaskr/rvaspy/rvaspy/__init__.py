from .version import get_version
from .client import HOST, CLIENT, RVASP


__version__ = get_version()


def connect(host=HOST, name=CLIENT):
    """
    Connect to the rVASP server at the specified host.

    Parameters
    ----------
    host : str, default="localhost:4434"
        Specify the host:port to connect the grpc insecure channel on.

    name : str, default="rvaspy"
        Optionally specify a client specific name for better server-side logging.
    """
    return RVASP(name=name, host=host)

# rVASPy

**Python rVASP implementation**

To send messages using the python API, make sure you can import the modules from `rvaspy` - the simplest way to do this is to install the package in editable mode as follows:

```
$ pip install -e ./lib/python/rvaspy
```

This will use the `setup.py` file in the `rvaspy` directory to install the package to your `$PYTHON_PATH`. Because it is in editable mode, any time you regenerate the protocol buffers or pull the repository, the module should be updated on the next time you import. An example script for using the package is as follows:

```python
import rvaspy

api = rvaspy.connect("localhost:4434")

cmds = [
    api.account_request("robert@bobvasp.co.uk"),
    api.transfer_request("robert@bobvasp.co.uk", "mary@alicevasp.us", 42.99)
]

for msg in api.stub.LiveUpdates(iter(cmds)):
    print(msg)
```

Note that the `RVASP` api client is not fully implemented yet.
import logging
import os

import jsons
from flask import Flask
from flask_cors import CORS
from flask_restful import Api
from flask_socketio import SocketIO

from flaskr import db
from flaskr.api.vasps import Vasps
from flaskr.models.user_wallet import UserWallet
from flaskr.models.vasp_details import VaspDetails
from flaskr.service.socket_manager import SocketManager
from flaskr.simulator.vasp_simulator import VaspSimulator


def create_app(test_config=None):
    # create and configure the app
    logging.basicConfig(level=logging.DEBUG)
    app = Flask(__name__, instance_relative_config=True)
    CORS(app)
    app.config.from_mapping(
        SECRET_KEY='dev',
        DATABASE=os.path.join(app.instance_path, 'vasps.sqlite')
    )

    db.init_app(app)

    logging.basicConfig(level=logging.DEBUG)

    print(jsons.dump(app.config))

    if test_config is None:
        # load the instance config, if it exists, when not testing
        app.config.from_pyfile('config.py', silent=True)
    else:
        # load the test config if passed in
        app.config.from_mapping(test_config)

    # ensure the instance folder exists
    try:
        os.makedirs(app.instance_path)
    except OSError:
        pass

    simulator = setup_simulator(app)
    setup_api(app)
    setup_websockets(app, simulator)

    return app


def setup_simulator(app: Flask):
    # TODO: config file tells us which simulator to launch
    vasp_details = VaspDetails('BOB-GUID', 'BobVASP', 'description', None, 'private-Key', 'public-key',
                               [
                                   UserWallet('ROBERT-GUID', '18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh'),
                                   UserWallet('AMY-GUID', 'amy@bobvasp')
                               ])
    return VaspSimulator(vasp_details)


def setup_api(app: Flask):
    api = Api(app)

    api.add_resource(Vasps, '/vasps')


def setup_websockets(app: Flask, simulator: VaspSimulator):
    socketio = SocketIO(app, cors_allowed_origins='*', logger=True, engineio_logger=True)
    socketio.run(app)

    socket_manager = SocketManager(socketio, simulator)


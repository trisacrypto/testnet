import io from 'socket.io-client';
import {VaspLogMessage} from "../models/VaspLogMessage";
import {VaspContext} from "../models/VaspContext";
import {Transaction} from "../models/Transaction";
import EventEmitter from "eventemitter3";
import { CryptoFormData } from '../models/CryptoFormData';
import {TransactionRequest} from "../models/TransactionRequest";

export class VaspStreamingService {
    vaspContext: VaspContext
    websocketServiceUrl: string
    webSocket: SocketIOClient.Socket
    vaspLogEmitter: EventEmitter<string, VaspLogMessage>
    transactionEmitter: EventEmitter<string, Transaction>

    constructor(vaspContext: VaspContext, websocketServiceUrl: string) {
        console.log("Creating websocket connection " + JSON.stringify(vaspContext) + ":" + websocketServiceUrl)
        this.vaspContext = vaspContext
        this.websocketServiceUrl = websocketServiceUrl

        this.webSocket = io.connect(websocketServiceUrl, {
            reconnection: false
        });

        this.vaspLogEmitter = new EventEmitter<string, VaspLogMessage>()
        this.transactionEmitter = new EventEmitter<string, Transaction>()

        this.setupWebsocket()
    }

    setupWebsocket() {
        this.webSocket.on('connect', () => {
            console.log('socket connected');

            if (this.vaspContext) {
                this.webSocket.emit('vasp_context', JSON.stringify(this.vaspContext))
            }
        })

        this.webSocket.on('vasp_log_message', (data: any) => {
            console.log('on vasp_log_message event data ' + data);

            let logMessage = JSON.parse(data) as VaspLogMessage

            if (logMessage) {
                this.vaspLogEmitter.emit('new', logMessage)
            }
        })

        this.webSocket.on('transaction', (data: any) => {
            console.log('on transaction event data ' + data);

            let transaction = JSON.parse(data) as Transaction
            if (transaction) {
                this.transactionEmitter.emit('new', transaction)
            }
        })


        this.webSocket.io.on('disconnect', () => {
            console.log('socket disconnected');
        });

    }

    sendTransactionRequest(transactionRequest: TransactionRequest) {
        this.webSocket.emit("transaction_request",
            JSON.stringify(transactionRequest))
    }
}

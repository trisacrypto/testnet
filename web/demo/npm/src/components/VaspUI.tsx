import React, {Component} from 'react'

import './Components.css';

import {VaspDetails} from "../models/VaspDetails";
import UserWalletSelector from "./UserWalletSelector";
import {UserWallet} from "../models/UserWallet";
import SendCryptoForm from "./SendCryptoForm";
import {CryptoFormData} from '../models/CryptoFormData';
import VaspLogConsole from "./VaspLogConsole";
import {VaspLogMessage} from "../models/VaspLogMessage";
import TransactionLedger from "./TransactionLedger";
import {Transaction} from "../models/Transaction";
import {VaspStreamingService} from "../services/VaspStreamingService";
import {VaspContext} from "../models/VaspContext";
import WalletDisplay from "./WalletDisplay";
import {Button, Card, Spin} from "antd";

interface VaspUIState {
    selectedUserWallet: UserWallet | undefined
    vaspMessages: VaspLogMessage[]
    transactions: Transaction[]
    isLoading: boolean
}

interface VaspUIProps {
    vaspContext: VaspContext
    vaspDetails: VaspDetails
    targetVasps: VaspDetails[]
    isOriginator: boolean
    onBeneficiaryVaspSelected: (vaspId: string) => void
    onVaspDSToShow: (vaspDetails: VaspDetails) => void
    onTransactionToShow: (transaction: Transaction) => void
    onStartSendSelected: (userWalletId: string) => void
    onClosePopup: () => void
    onCloseUI: (isOriginator: boolean) => void
    vaspStreamingService: VaspStreamingService
}

class VaspUI extends Component<VaspUIProps, VaspUIState> {
    constructor(props: VaspUIProps) {
        super(props);

        console.log("Created vasp UI " + props)

        this.state = {
            selectedUserWallet: undefined,
            vaspMessages: [
                {
                    vasp_id: this.props.vaspDetails.vasp_id,
                    timestamp: "",
                    message: "",
                    message_unencrypted: "",
                    color_code: "ffffff"
                }
            ],
            transactions: [],
            isLoading: true
        }

        this.props.vaspStreamingService.vaspLogEmitter.on('new', (vaspMessage: VaspLogMessage) => {
            this.state.vaspMessages.push(vaspMessage)
            this.setState({
                vaspMessages: this.state.vaspMessages
            })
        })
        this.props.vaspStreamingService.transactionEmitter.on('new', (transaction: Transaction) => {
            this.state.transactions.push(transaction)
            this.setState({
                transactions: this.state.transactions
            })
        })
    }

    async componentDidMount() {
        function timeout(delay: number) {
            return new Promise( res => setTimeout(res, delay) );
        }

        await timeout(1500);

        this.setState({
            isLoading: false,
            vaspMessages: [
                {
                    vasp_id: this.props.vaspDetails.vasp_id,
                    timestamp: "",
                    message: "",
                    message_unencrypted: "cmd> tail -f /var/log/vasp-exchange.log",
                    color_code: "ffffff"
                }
            ]
        })
    }

    onCloseSelected = () => {
        this.props.onCloseUI(this.props.isOriginator)
    }

    onUserWalletSelected = (userWalletId: string) => {
        console.log("Selected user wallet " + userWalletId)

        let selectedUser = this.props.vaspDetails.user_wallets
            .filter(user_wallet => user_wallet.user_wallet_id === userWalletId)[0]
        this.setState({selectedUserWallet: selectedUser})
    }

    onSendSelected = (cryptoFormData: CryptoFormData) => {

        if (this.state.selectedUserWallet == undefined) {
            alert("Select an originating wallet for transaction.")
        } else {

            // TODO: send transaction
            console.log("Send selected " + JSON.stringify(cryptoFormData))
        }
    }

    onDSSelected = () => {
        this.props.onVaspDSToShow(this.props.vaspDetails)
    }

    render() {
        var directoryServiceDetails =
            <h4><img src="trisa_x.png"/> Not registered with TRISA Directory Service</h4>

        if (this.props.vaspDetails.trisa_ds_entry) {
            directoryServiceDetails = <h4><img src="trisa_check.png"/> Registered with TRISA Directory Service <Button
                onClick={this.onDSSelected}>Details</Button></h4>
        }

        var walletDisplay

        if (this.state.isLoading) {
            walletDisplay = <WalletDisplay
                userWallets={[]}
                isOriginatingVasp={this.props.isOriginator}
                onSendSelected={this.props.onStartSendSelected}
            />
        } else {
            walletDisplay = <WalletDisplay
                userWallets={this.props.vaspDetails.user_wallets}
                isOriginatingVasp={this.props.isOriginator}
                onSendSelected={this.props.onStartSendSelected}
            />
        }
        var card = <Card title={"Admin Console: " + this.props.vaspDetails.display_name}
                     extra={<a href="#" onClick={this.onCloseSelected}><img src={'close.png'}/></a>}
                     className='admin-console-container'>
            <div>
                {directoryServiceDetails}
                {walletDisplay}
                <h4>VASP Log Console</h4>
                <VaspLogConsole
                    vaspDetails={this.props.vaspDetails}
                    vaspMessages={this.state.vaspMessages}
                />
                <h4>Transaction Ledger</h4>
                <TransactionLedger
                    currentVaspId={this.props.vaspDetails.vasp_id}
                    transactions={this.state.transactions}
                    onTransactionSelected={this.props.onTransactionToShow}/>
            </div>
        </Card>

        if (this.state.isLoading) {
            let loadingTip = "Loading Admin Console for " + this.props.vaspDetails.display_name
            return <Spin tip={loadingTip}>
                {card}
            </Spin>
        } else {
            return card
        }

    }

}

export default VaspUI
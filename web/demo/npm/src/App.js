import React, {Component} from 'react';
import './App.less';
import './components/Components.css';
import {VaspRestService} from "./services/VaspRestService";
import VaspSelector from "./components/VaspSelector";
import {Button, Card, Col, Row, Space} from "antd";
import VaspUI from "./components/VaspUI";
import {v4 as uuidv4} from 'uuid';
import {Transaction} from "./models/Transaction";
import TrisaDSCard from "./components/TrisaDSCard";
import TransactionDetailsCard from "./components/TransactionDetailsCard";
import SendCryptoPopup from "./components/SendCryptoPopup";
import {CryptoFormData} from "./models/CryptoFormData";
import {VaspStreamingService} from "./services/VaspStreamingService";

class App extends Component {

    vaspService = new VaspRestService('https://demo.bob.vaspbot.net')
    sessionId = uuidv4();

    state = {
        vasps: [],
        selectedOriginatingVasp: null,
        selectedBeneficiaryVasp: null,
        originatingVaspStreamingService: null,
        beneficiaryVaspStreamingService: null,
        targetVasps: [],
        showVaspDsCardPopup: false,
        vaspDSToShow: null,
        showTransactionPopup: false,
        transactionToShow: null,
        selectedOriginatingWallet: null,
        showSendCryptoFormPopup: false
    }

    componentDidMount() {
        this.vaspService
            .getVasps()
            .then(data => {
                this.setState({vasps: data})
            })
            .catch(console.log)

    }

    onShowCryptoFormSelected = (originatingWallet) => {
        if (this.state.selectedBeneficiaryVasp != null) {
            this.setState({
                selectedOriginatingWallet: originatingWallet,
                showSendCryptoFormPopup: true
            })
        } else {
            alert("Select a beneficiary VASP to start transaction")
        }
    }
    onOriginatingVaspSelected = (vaspId) => {
        let selectedVasp = this.state.vasps
            .filter(vasp => vasp.vasp_id === vaspId)[0]

        let targetVasps = this.state.vasps
            .filter(vasp => vasp.vasp_id !== vaspId)

        this.setState({
            selectedOriginatingVasp: selectedVasp,
            selectedBeneficiaryVasp: null,
            targetVasps: targetVasps,
            originatingVaspStreamingService: new VaspStreamingService({'vasp_id': selectedVasp.vasp_id, 'context_id': this.sessionId, originator: true}, 'https://demo.bob.vaspbot.net'),
            beneficiaryVaspStreamingService: null
        })
    }

    onCloseVaspUI = (isOriginator) => {
        if (isOriginator) {
            this.onOriginatingVaspCleared()
        } else {
            this.onBeneficiaryVaspCleared()
        }
    }

    onOriginatingVaspCleared = () => {
        this.setState({
            selectedOriginatingVasp: null,
            selectedBeneficiaryVasp: null,
            originatingVaspStreamingService: null,
            beneficiaryVaspStreamingService: null,
            targetVasps: []
        })
    }

    onBeneficiaryVaspSelected = (vaspId) => {
        console.log("Selected beneficiary vasp " + vaspId)

        let selectedVasp = this.state.vasps
            .filter(vasp => vasp.vasp_id === vaspId)[0]

        console.log("Selected beneficiary vasp " + JSON.stringify(selectedVasp))
        this.setState({
            selectedBeneficiaryVasp: selectedVasp,
            beneficiaryVaspStreamingService: new VaspStreamingService({'vasp_id': selectedVasp.vasp_id, 'context_id': this.sessionId, originator: false}, 'https://demo.bob.vaspbot.net')
        })
    }

    onBeneficiaryVaspCleared = () => {
        this.setState({
            selectedBeneficiaryVasp: null,
            beneficiaryVaspStreamingService: null
        })
    }


    closePopups = () => {
        this.setState({
            showTransactionPopup: false,
            showVaspDsCardPopup: false,
            showSendCryptoFormPopup: false,
            selectedOriginatingWallet: null,
            transactionToShow: null,
            vaspDSToShow: null

        });
    }

    showTransactionPopup = (transaction) => {
        this.setState({
            transactionToShow: transaction,
            showTransactionPopup: true
        });
    }

    showVaspDSPopup = (vaspDetails) => {
        this.setState({
            vaspDSToShow: vaspDetails,
            showVaspDsCardPopup: true
        });
    }

    onCloseCryptoPopup = () => {
        this.setState({
            showSendCryptoFormPopup: false,
            selectedOriginatingWallet: null
        })
    }
    onSendSelected = (cryptoFormData) => {
        console.log("Send selected " + JSON.stringify(cryptoFormData))
        this.setState({
            showSendCryptoFormPopup: false
        })

        let transactionRequest = {
            context_id: this.sessionId,
            originator_vasp_id: this.state.selectedOriginatingVasp.vasp_id,
            originator_wallet_id: this.state.selectedOriginatingWallet,
            beneficiary_vasp_id: this.state.selectedBeneficiaryVasp.vasp_id,
            beneficiary_wallet_id: cryptoFormData.beneficiaryWalletId,
            crypto_type: cryptoFormData.cryptoType,
            amount: cryptoFormData.amount
        }

        // start the transaction
        this.state.originatingVaspStreamingService.sendTransactionRequest(transactionRequest)
    }

    render() {


        var leftPane = null

        var topRow = <div className="top-row" align="center"><h2>VASP Cryptocurrency Exchange Simulator utilizing TRISA protocol</h2><h3>Learn more about integrating TRISA: <a target="_blank" href="https://trisatest.net">Documentation</a></h3></div>

        if (this.state.selectedOriginatingVasp) {
            console.log("Originating vasp creation")

            let originatingTitle = "Originating VASP: " + this.state.selectedOriginatingVasp.display_name
            leftPane = <div align="left" className="vasp-pane">

                <VaspUI
                    vaspContext={{'vasp_id': this.state.selectedOriginatingVasp.vasp_id, 'context_id': this.sessionId}}
                    vaspDetails={this.state.selectedOriginatingVasp}
                    targetVasps={this.state.targetVasps}
                    isOriginator={true}
                    onBeneficiaryVaspSelected={this.onBeneficiaryVaspSelected}
                    onVaspDSToShow={this.showVaspDSPopup}
                    onTransactionToShow={this.showTransactionPopup}
                    onStartSendSelected={this.onShowCryptoFormSelected}
                    onClosePopup={this.closePopups}
                    onCloseUI={this.onCloseVaspUI}
                    vaspStreamingService={this.state.originatingVaspStreamingService}
                />
            </div>
        } else {
            leftPane = <div align="center" style={{padding: "150px"}}>
                <h3>Select the originating VASP</h3>
                <VaspSelector vaspDetails={this.state.vasps}
                              selectedVasp={this.state.selectedOriginatingVasp}
                              onSelect={this.onOriginatingVaspSelected}/>
            </div>
        }

        var rightPane = ''

        if (this.state.selectedBeneficiaryVasp) {
            let beneficiaryTitle = "Beneficiary VASP: " + this.state.selectedBeneficiaryVasp.display_name
            rightPane = <div align="left" className="vasp-pane">
                <VaspUI
                    vaspContext={{'vasp_id': this.state.selectedBeneficiaryVasp.vasp_id, 'context_id': this.sessionId}}
                    vaspDetails={this.state.selectedBeneficiaryVasp}
                    targetVasps={[]}
                    isOriginator={false}
                    onVaspDSToShow={this.showVaspDSPopup}
                    onTransactionToShow={this.showTransactionPopup}
                    onStartSendSelected={this.onShowCryptoFormSelected}
                    onClosePopup={this.closePopups}
                    onCloseUI={this.onCloseVaspUI}
                    vaspStreamingService={this.state.beneficiaryVaspStreamingService}
                />

            </div>
        } else if (this.state.selectedOriginatingVasp) {
            rightPane = <div align="center" style={{padding: "150px"}}>
                <h3>Select the beneficiary VASP</h3>
                <VaspSelector vaspDetails={this.state.targetVasps}
                              selectedVasp={this.state.selectedBeneficiaryVasp}
                              onSelect={this.onBeneficiaryVaspSelected}/>
            </div>


        }

        let dsPopup = this.state.showVaspDsCardPopup ?
            <TrisaDSCard
                vaspDetails={this.state.vaspDSToShow}
                onClose={this.closePopups}
            />
            : null

        let sendCryptoPopup = this.state.showSendCryptoFormPopup ?
            <SendCryptoPopup targetVasp={this.state.selectedBeneficiaryVasp}
                             originatingWallet={this.state.selectedOriginatingWallet}
                             originatingVasp={this.state.selectedOriginatingVasp}
                             availableCurrencies={["BTC", "ETH"]}
                             onSendSelected={this.onSendSelected}
                             onClose={this.onCloseCryptoPopup}
            />
        : null

        let transactionPopup = this.state.showTransactionPopup ?
            <TransactionDetailsCard
                transaction={this.state.transactionToShow}
                onClose={this.closePopups}
            />
            : null


        console.log("Rendering app")

        var style = null
        if (this.state.selectedOriginatingVasp && this.state.selectedBeneficiaryVasp) {
            style = {
                height:"100%",
                backgroundImage: "url(send_crypto_vertical.png)",
                backgroundRepeat: "no-repeat",
                backgroundAttachment: "fixed",
                backgroundPosition: "center"
            }
        } else {
            style = {height:"100%"}
        }
        return (
            <div style={style}>
                {topRow}
                <Row gutter={[16, 16]} style={{height:"100%"}}>
                    <Col className="gutter-row" span={12}  style={{height:"100%"}}>
                        <div className="gutter-row">
                            {leftPane}
                        </div>
                    </Col>
                    <Col className="gutter-row" span={12}  style={{height:"100%"}}>
                        <div className="gutter-row">
                            {rightPane}
                        </div>
                    </Col>
                </Row>
                <div>
                    {transactionPopup}
                </div>
                <div>
                    {sendCryptoPopup}
                </div>
                <div>
                    {dsPopup}
                </div>
            </div>
        )
    }
}

export default App;

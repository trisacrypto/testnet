import React, {Component} from 'react'
import {AutoComplete, Button, Card, Col, Row} from 'antd';

import {VaspDetails} from "../models/VaspDetails";
import {CryptoFormData} from '../models/CryptoFormData';
import NumericInput from "./NumericInput";
import CurrencySelector from "./CurrencySelector";

interface SendCurrencyState {
    selectedBeneficiary: string | undefined
    selectedCurrency: string | undefined
    selectedAmount: string
}

interface SendCurrencyProps {
    targetVasp: VaspDetails;
    originatingWallet: string;
    originatingVasp: VaspDetails;
    availableCurrencies: string[];
    onSendSelected: (formData: CryptoFormData) => void;
    onClose: () => void;
}

class SendCryptoPopup extends Component<SendCurrencyProps, SendCurrencyState> {


    constructor(props: SendCurrencyProps) {
        super(props);

        this.state = {
            selectedCurrency: 'BTC',
            selectedAmount: '0.1',
            selectedBeneficiary: undefined
        }
    }

    onBeneficiaryUpdated = (beneficiary: string) => {
        this.setState({selectedBeneficiary: beneficiary})
    }

    onAmountChange = (amount: string) => {
        this.setState({selectedAmount: amount});
    }

    onCurrencySelected = (currency: string) => {
        this.setState({selectedCurrency: currency})
    }

    handleValidation = () => {
        let selectedAmount = parseFloat(this.state.selectedAmount)

        console.log(selectedAmount + ":amount")
        if (isNaN(selectedAmount) || selectedAmount <= 0) {
            alert("Entry transaction amount")
        } else if (this.state.selectedBeneficiary == undefined) {
            alert("Select a beneficiary wallet")
        } else if (this.state.selectedCurrency == undefined) {
            alert("Select a currency type")
        } else {
            this.props.onSendSelected({
                beneficiaryVaspId: this.props.targetVasp.vasp_id,
                beneficiaryWalletId: this.state.selectedBeneficiary,
                cryptoType: this.state.selectedCurrency,
                amount: selectedAmount
            })
        }

    }

    render() {

        let options = this.props.targetVasp.user_wallets.map(
            user => {
                return {
                    label: user.wallet_address,
                    value: user.wallet_address
                }
            }
        )

        return <div className='popup'>
            <div className='popup_inner'>
                <Card title="Start Transaction"
                      extra={<a href="#" onClick={this.props.onClose}><img src={'close.png'}/></a>}>

                    <Row>
                        <Col className="card-left-table">
                            <h3>Originating VASP:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.originatingVasp.display_name}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Originating Wallet:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.originatingWallet}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Beneficiary VASP:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.targetVasp.display_name}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Beneficiary Wallet:</h3>
                        </Col>
                        <Col className="card-right-table">
                            <AutoComplete
                                className="autocomplete-full-width"
                                options={options}
                                onChange={(value) => this.onBeneficiaryUpdated(value)}
                                placeholder="Select Beneficiary Wallet"
                            />
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Amount:</h3>
                        </Col>
                        <Col className="card-right-table">
                            <NumericInput value={"" + this.state.selectedAmount} onChange={this.onAmountChange}/>
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Cryptocurrency Type:</h3>
                        </Col>
                        <Col className="card-right-table">
                            <CurrencySelector
                                availableCurrencies={this.props.availableCurrencies}
                                selectedCurrency={this.state.selectedCurrency}
                                onSelect={this.onCurrencySelected}
                            />
                        </Col>
                    </Row>
                    <Row className="card-center-row" justify="center">
                        <Button onClick={this.handleValidation}>
                            Send
                        </Button>
                    </Row>
                </Card>
            </div>
        </div>
    }

}

export default SendCryptoPopup;
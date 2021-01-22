import React, {Component} from 'react'
import {Button, Col, Input, Row} from 'antd';

import {VaspDetails} from "../models/VaspDetails";
import VaspSelector from "./VaspSelector";
import {CryptoFormData} from '../models/CryptoFormData';
import NumericInput from "./NumericInput";
import CurrencySelector from "./CurrencySelector";

interface SendCurrencyState {
    selectedBeneficiaryVasp: VaspDetails | undefined
    selectedBeneficiary: string | undefined
    selectedCurrency: string | undefined
    selectedAmount: string
}

interface SendCurrencyProps {
    targetVasps: VaspDetails[];
    availableCurrencies: string[];
    onBeneficiaryVaspSelected: (vaspId: string) => void;
    onSendSelected: (formData: CryptoFormData) => void;
}

class SendCurrencyForm extends Component<SendCurrencyProps, SendCurrencyState> {


    constructor(props: SendCurrencyProps) {
        super(props);

        this.state = {
            selectedBeneficiaryVasp: undefined,
            selectedCurrency: undefined,
            selectedAmount: '',
            selectedBeneficiary: undefined
        }
    }

    onBeneficiaryUpdated = (beneficiary: string) => {
        this.setState({selectedBeneficiary: beneficiary})
    }
    onBeneficiaryVaspSelected = (vaspId: string) => {
        console.log("Selected beneficiary vasp " + vaspId)

        let selectedVasp = this.props.targetVasps
            .filter(vasp => vasp.vasp_id === vaspId)[0]

        this.setState({selectedBeneficiaryVasp: selectedVasp})

        this.props.onBeneficiaryVaspSelected(selectedVasp.vasp_id)
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
        if (this.state.selectedBeneficiaryVasp == undefined) {
            alert("Select a beneficiary vasp")
        } else if (isNaN(selectedAmount) || selectedAmount <= 0) {
            alert("Entry transaction amount")
        } else if (this.state.selectedBeneficiary == undefined) {
            alert("Select a beneficiary wallet")
        } else if (this.state.selectedCurrency == undefined) {
            alert("Select a currency type")
        } else {
            this.props.onSendSelected({
                beneficiaryVaspId: this.state.selectedBeneficiaryVasp.vasp_id,
                beneficiaryWalletId: this.state.selectedBeneficiary,
                cryptoType: this.state.selectedCurrency,
                amount: selectedAmount
            })
        }

    }

    render() {

        return <div>
            <Row gutter={[16, 16]}>
                <Col span={12}>
                    <Input placeholder="Beneficiary wallet address"
                           onChange={({target}) => this.onBeneficiaryUpdated(target.value)}/>
                </Col>
                <Col span={12}>
                    <VaspSelector vaspDetails={this.props.targetVasps} selectedVasp={this.state.selectedBeneficiaryVasp}
                                  onSelect={this.onBeneficiaryVaspSelected}/>
                </Col>
            </Row>
            <Row gutter={[16, 16]}>
                <Col span={10}>
                    <NumericInput value={"" + this.state.selectedAmount} onChange={this.onAmountChange}/>
                </Col>
                <Col span={10}>

                    <CurrencySelector
                        availableCurrencies={this.props.availableCurrencies}
                        selectedCurrency={this.state.selectedCurrency}
                        onSelect={this.onCurrencySelected}
                    />
                </Col>
                <Col span={4}>

                    <Button onClick={this.handleValidation}>
                        Send
                    </Button>
                </Col>
            </Row>

        </div>
    }

}

export default SendCurrencyForm;
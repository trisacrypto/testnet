import React, {Component} from 'react'
import {Button, Card, Col, Row, Typography} from "antd";
import {Transaction} from "../models/Transaction";
import {formatTransactionTime} from "./TransactionLedger";
import TextArea from "antd/lib/input/TextArea";

interface TransactionDetailsCardProps {
    transaction: Transaction,
    onClose: () => void;
}

class TransactionDetailsCard extends Component<TransactionDetailsCardProps> {

    render() {
        return <div className='popup'>
            <div className='popup_inner'>
                <Card title="Transaction Details" extra={<a href="#" onClick={this.props.onClose}><img src={'close.png'}/></a>}>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Transaction ID:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.transaction.transaction_id}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Timestamp:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {formatTransactionTime(this.props.transaction)}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Originating Wallet:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.transaction.originating_wallet}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Originating VASP:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.transaction.originating_vasp_display_name}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Beneficiary Wallet:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.transaction.beneficiary_wallet}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>Beneficiary VASP:</h3>
                        </Col>
                        <Col className="card-right-table">
                            {this.props.transaction.beneficiary_vasp_display_name}
                        </Col>
                    </Row>
                    <Row>
                        <Col className="card-left-table">
                            <h3>IVMS101:</h3>
                        </Col>
                        <Col className="card-right-table">
                            <TextArea readOnly={true} rows={6} defaultValue={JSON.stringify(JSON.parse(this.props.transaction.ivms101Data), null, 2)}/>
                        </Col>
                    </Row>

                </Card>
            </div>
        </div>;
    }

};


export default TransactionDetailsCard;
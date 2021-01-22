import React, {Component} from 'react'
import {VaspDetails} from "../models/VaspDetails";
import {Button, Card, Col, Row, Typography} from "antd";

interface TrisaDSCardProps {
    vaspDetails: VaspDetails,
    onClose: () => void;
}

class TrisaDSCard extends Component<TrisaDSCardProps> {

    render() {
        var details = <div className='popup'>
            <div className='popup_inner'>
                <b>VASP {this.props.vaspDetails.display_name} is not registered in TRISA Directory Service.</b>
                <Button onClick={this.props.onClose}>Close</Button>
            </div>
        </div>

        if (this.props.vaspDetails.trisa_ds_entry) {
            let title = "TRISA Directory Service Entry: " + this.props.vaspDetails.display_name
            details =
                <div className='popup'>
                    <div className='popup_inner'>
                        <Card title={title} extra={<a href="#" onClick={this.props.onClose}><img src={'close.png'}/></a>}>
                            <Row>
                                <Col className="card-left-table">
                                    <h3>TRISA DS Status:</h3>
                                </Col>
                                <Col className="card-right-table">
                                    <img height="20px" src="trisa_check.png"/> Verified
                                </Col>
                            </Row>
                            <Row>
                                <Col className="card-left-table">
                                    <h3>TRISA DS ID:</h3>
                                </Col>
                                <Col className="card-right-table">
                                    {this.props.vaspDetails.trisa_ds_entry.trisa_ds_id}
                                </Col>
                            </Row>
                            <Row>
                                <Col className="card-left-table">
                                    <h3>TRISA DS Name:</h3>
                                </Col>
                                <Col className="card-right-table">
                                    {this.props.vaspDetails.trisa_ds_entry.display_name}
                                </Col>
                            </Row>
                            <Row>
                                <Col className="card-left-table">
                                    <h3>TRISA P2P Protocol URL:</h3>
                                </Col>
                                <Col className="card-right-table">
                                    {this.props.vaspDetails.trisa_ds_entry.trisa_protocol_host}
                                </Col>
                            </Row>
                            <Row>
                                <Col className="card-left-table">
                                    <h3>Public Key:</h3>
                                </Col>
                                <Col className="card-right-table">
                                    <Typography.Text ellipsis={true} className="card-wrap">
                                        {this.props.vaspDetails.public_key}
                                    </Typography.Text>
                                </Col>
                            </Row>

                        </Card>
                    </div>
                </div>
        }

        return details;
    }

};

export default TrisaDSCard;
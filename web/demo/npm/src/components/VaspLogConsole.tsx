import React, {Component} from 'react'
import {VaspLogMessage} from "../models/VaspLogMessage";
import {List, Row} from "antd";
import './Components.css';
import {VaspDetails} from "../models/VaspDetails";


interface VaspConsoleLogProps {
    vaspDetails: VaspDetails
    vaspMessages: VaspLogMessage[]
}

class VaspLogConsole extends Component<VaspConsoleLogProps> {

    getListName(): string {
        return "console-" + this.props.vaspDetails.vasp_id
    }
    // Check the change of the list, and trigger the scroll
    componentDidUpdate(prevProps: any, prevState: any) {
        window.document.getElementsByClassName(
            this.getListName()
        )[0].scrollIntoView({block: "end"})
    }

    render() {
        return <div className="vasp-console-container">

            <List
                className={this.getListName()}
                dataSource={this.props.vaspMessages}
                pagination={false}
                itemLayout="vertical"
                size="small"
                bordered={false}
                renderItem={item => (
                    <Row
                        style={
                            {
                                color: '#' + item.color_code,
                                padding: '0px 0px',
                                borderBottom: '0px',
                            }
                        }
                    >{formatVaspLogMessage(item)}</Row>
                )}
            /></div>
    }
}


function formatVaspLogMessage(vaspMessage: VaspLogMessage): string {
    return vaspMessage.timestamp + ' ' + vaspMessage.message_unencrypted
    // if (vaspMessage.timestamp == 0) {
    //     // system message
    //     return '' + vaspMessage.message_unencrypted
    // }
    //
    // let date = new Date(vaspMessage.timestamp)
    //
    // console.log("Date is " + date + ":" + vaspMessage.timestamp)
    //
    // let dateOptions = {day: '2-digit', month: '2-digit', year: 'numeric'}
    // let timeOptions = {hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit'}
    //
    // let dateFormat = new Intl.DateTimeFormat([], dateOptions).format(date)
    // let timeFormat = new Intl.DateTimeFormat([], timeOptions).format(date)
    //
    // return dateFormat + ' ' + timeFormat + ' ' + vaspMessage.message_unencrypted
}

export default VaspLogConsole;
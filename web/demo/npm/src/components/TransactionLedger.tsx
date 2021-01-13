import React from 'react'
import {Table} from "antd";
import './Components.css';
import {Transaction} from "../models/Transaction";


const TransactionLedger: React.FunctionComponent<{
    currentVaspId: string
    transactions: Transaction[]
    onTransactionSelected: (transaction: Transaction) => void
}> = (props) => {

    let columns = [
        {
            title: 'Timestamp',
            dataIndex: 'timestamp',
            key: 'timestamp'
        },
        {
            title: 'Transaction ID',
            dataIndex: 'transactionId',
            key: 'transactionId'
        },
        {
            title: 'Direction',
            dataIndex: 'direction',
            key: 'direction'
        },
        {
            title: 'Details',
            dataIndex: 'details',
            key: 'details'
        }
    ]

    let dataSource = props.transactions.map(transaction => {
        return {
            'timestamp': formatTransactionTime(transaction),
            'transactionId': transaction.transaction_id,
            'direction': formatTransactionDirection(transaction, props.currentVaspId),
            'details': formatTransactionDetails(transaction)
        }
    })

    function handleRow(transactionId: String) {

        let selectedTransaction = props.transactions
            .filter(transaction => transaction.transaction_id === transactionId)[0]

        props.onTransactionSelected(selectedTransaction)

    }

    return <div className="transaction-ledger-container">
        <Table
            dataSource={dataSource}
            columns={columns}
            pagination={false}
            size="small"
            bordered={false}
            scroll={{y: 200}}
            onRow={(record, rowIndex) => {
                return {
                    onClick: event => {
                        handleRow(record.transactionId)
                    }
                };
            }}
        /></div>
}


export function formatTransactionTime(transaction: Transaction): string {
    return transaction.timestamp

    // if (transaction.timestamp == 0) {
    //     return ''
    // }
    // let date = new Date(transaction.timestamp)
    //
    // let dateOptions = {day: '2-digit', month: '2-digit', year: 'numeric'}
    // let timeOptions = {hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit'}
    //
    // let dateFormat = new Intl.DateTimeFormat([], dateOptions).format(date)
    // let timeFormat = new Intl.DateTimeFormat([], timeOptions).format(date)
    //
    // return dateFormat + ' ' + timeFormat
}

function formatTransactionDirection(transaction: Transaction, current_vasp_id: string): string {
    let direction = current_vasp_id === transaction.beneficiary_vasp_id ? "Incoming" : "Outgoing"
    return direction
}

function formatTransactionDetails(transaction: Transaction): string {
    return transaction.originating_wallet + ' => ' + transaction.beneficiary_wallet
}


export default TransactionLedger;
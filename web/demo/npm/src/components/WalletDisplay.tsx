import React from 'react'
import {Table} from 'antd';

import {UserWallet} from "../models/UserWallet";

const WalletDisplay: React.FunctionComponent<{
    userWallets: UserWallet[];
    isOriginatingVasp: boolean;
    onSendSelected: (userWalletId: string) => void;
}> = (props) => {

    var action = ''
    if (props.isOriginatingVasp) {
        action = "Start transaction from wallet"
    }
    let columns = [
        {
            title: 'Wallets hosted by this VASP',
            dataIndex: 'wallet',
            key: 'wallet'
        },
        {
            title: '',
            key: 'action',
            render: (text: any, record: any) => (
                <a href="#" onClick={() => props.onSendSelected(record.wallet)}>{action}</a>
            ),
        }
    ]

    let dataSource = props.userWallets.map(userWallet => {
        return {
            'wallet': userWallet.wallet_address,
        }
    })
    return <div className="transaction-ledger-container">
        <Table
            dataSource={dataSource}
            columns={columns}
            pagination={false}
            size="small"
            bordered={false}
            scroll={{y: 150}}
        /></div>
};

export default WalletDisplay;
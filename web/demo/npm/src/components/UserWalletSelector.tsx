import React from 'react'
import {Button, Dropdown, Menu} from 'antd';
import {DownOutlined, WalletOutlined} from '@ant-design/icons';

import {UserWallet} from "../models/UserWallet";

const UserWalletSelector: React.FunctionComponent<{
    userWallets: UserWallet[];
    selectedUserWallet: UserWallet | undefined;
    onSelect: (userWalletId: string) => void;
}> = (props) => {

    let menuItems = props.userWallets.map(userWallet => {
        return <Menu.Item key={userWallet.user_wallet_id} icon={<WalletOutlined/>}>
            {userWallet.wallet_address}
        </Menu.Item>
    });

    let menu = <Menu
        onClick={({key}) => props.onSelect("" + key)}
    >
        {menuItems}
    </Menu>

    return <Dropdown overlay={menu}>
        <Button>
            {props.selectedUserWallet ? props.selectedUserWallet.wallet_address : "Select User"} <DownOutlined/>
        </Button>
    </Dropdown>

};

export default UserWalletSelector;
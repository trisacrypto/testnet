import React from 'react'
import {Button, Dropdown, Menu} from 'antd';
import {DeploymentUnitOutlined, DownOutlined} from '@ant-design/icons';

const CurrencySelector: React.FunctionComponent<{
    availableCurrencies: string[];
    selectedCurrency: string | undefined;
    onSelect: (currency: string) => void;
}> = (props) => {

    let menuItems = props.availableCurrencies.map(currency => {
        return <Menu.Item key={currency} icon={<DeploymentUnitOutlined/>}>
            {currency}
        </Menu.Item>
    });

    let menu = <Menu
        onClick={({key}) => props.onSelect("" + key)}
    >
        {menuItems}
    </Menu>

    return <Dropdown overlay={menu}>
        <Button>
            {props.selectedCurrency ? props.selectedCurrency : "Select Crypto"} <DownOutlined/>
        </Button>
    </Dropdown>
}

export default CurrencySelector;
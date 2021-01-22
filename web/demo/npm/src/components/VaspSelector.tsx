import React from 'react'
import {Button, Dropdown, Menu} from 'antd';
import {CloudServerOutlined, DownOutlined} from '@ant-design/icons';

import {VaspDetails} from "../models/VaspDetails";

const VaspSelector: React.FunctionComponent<{
    vaspDetails: VaspDetails[];
    selectedVasp: VaspDetails | undefined;
    onSelect: (vaspId: string) => void;
}> = (props) => {

    let menuItems = props.vaspDetails.map(vasp => {
        return <Menu.Item key={vasp.vasp_id} icon={<CloudServerOutlined/>}>
            {vasp.display_name} - {vasp.description}
        </Menu.Item>
    });

    let menu = <Menu
        onClick={({key}) => props.onSelect("" + key)}
    >
        {menuItems}
    </Menu>

    return <Dropdown overlay={menu}>
        <Button>
            {props.selectedVasp ? props.selectedVasp.display_name : "Select VASP"} <DownOutlined/>
        </Button>
    </Dropdown>

};

export default VaspSelector;
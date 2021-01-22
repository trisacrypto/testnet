import React from 'react'
import {VaspDetails} from "../models/VaspDetails";

const VaspDisplayCard: React.FunctionComponent<{
    vaspDetails: VaspDetails;
}> = (props) => {

    let wallets = props.vaspDetails.user_wallets.map(wallet => {
        return <li key={wallet.user_wallet_id}>{wallet.wallet_address}</li>
    });

    return <div key={props.vaspDetails.vasp_id}>
        <h1>{props.vaspDetails.display_name}</h1>
        <h2>{props.vaspDetails.trisa_ds_entry}</h2>
        <h2>{props.vaspDetails.private_key}</h2>
        <h2>{props.vaspDetails.public_key}</h2>
        <h2>{props.vaspDetails.vasp_id}</h2>
        <ul>{wallets}</ul>
    </div>
        ;
};

export default VaspDisplayCard;
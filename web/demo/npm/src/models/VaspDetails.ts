import {TrisaDsEntry} from "./TrisaDsEntry";
import {UserWallet} from "./UserWallet";

export interface VaspDetails {
    vasp_id: string;
    display_name: string;
    description: string;
    trisa_ds_entry: TrisaDsEntry | null;
    private_key: string;
    public_key: string;
    user_wallets: UserWallet[];
}

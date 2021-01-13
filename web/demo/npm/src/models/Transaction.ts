export interface Transaction {
    timestamp: string;
    transaction_id: string;
    originating_wallet: string;
    originating_vasp_id: string;
    originating_vasp_display_name: string;
    beneficiary_wallet: string;
    beneficiary_vasp_id: string;
    beneficiary_vasp_display_name: string;
    ivms101Data: string;
}

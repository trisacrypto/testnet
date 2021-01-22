export interface TransactionRequest {
    context_id: string;
    originator_vasp_id: string;
    originator_wallet_id: string;
    beneficiary_vasp_id: string;
    beneficiary_wallet_id: string;
    crypto_type: string;
    amount: number;
}
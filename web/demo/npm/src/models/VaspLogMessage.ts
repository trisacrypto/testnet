export interface VaspLogMessage {
    vasp_id: string;
    timestamp: string;
    message: string;
    message_unencrypted: string;
    color_code: string;
}

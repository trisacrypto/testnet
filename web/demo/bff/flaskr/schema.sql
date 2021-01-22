DROP TABLE IF EXISTS vasps;
DROP TABLE IF EXISTS wallets;

CREATE TABLE vasps (
  vasp_id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  description TEXT NOT NULL,
  private_key TEXT NOT NULL,
  public_key TEXT NOT NULL,
  websocket_address TEXT NOT NULL,
  trisa_ds_id TEXT,
  trisa_ds_name TEXT,
  trisa_protocol_host TEXT
);

CREATE TABLE wallets (
  wallet_id TEXT PRIMARY KEY,
  vasp_id TEXT NOT NULL,
  wallet_address TEXT NOT NULL,
  FOREIGN KEY (vasp_id) REFERENCES vasps (vasp_id)
);
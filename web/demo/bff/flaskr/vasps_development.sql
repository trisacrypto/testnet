INSERT INTO vasps(vasp_id, display_name, description, private_key, public_key, websocket_address, trisa_ds_id, trisa_ds_name, trisa_protocol_host)
 VALUES('api.bob.vaspbot.net', 'Bob VASP', 'Run of the mill VASP.  Registered in the Trisa Directory Service.', 'private', 'MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCAmB2ZjvOkpiwOMQCyMnUhQipXYlGhxI673WHVXWcA0MYsxdksF426UTY+Lx+SvQjIH0B2BS5O9WmiRZcPD8csly0DoOen8QiM8ZIRt8pW98V85GFZjlfWGF2ML0HgxSHE6g+9UfJPH9p6uH5TGKWBBGpzBMx44L4t9zyHJ2lVMwIDAQAB', 'admin.bob.vaspbot.net:443', '1', 'Bobs Friendly VASP', 'http://124.52.4.63:4443');

INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('robert@bobvasp.co.uk', 'api.bob.vaspbot.net', '18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('george@bobvasp.co.uk', 'api.bob.vaspbot.net', '1LgtLYkpaXhHDu1Ngh7x9fcBs5KuThbSzw');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('larry@bobvasp.co.uk', 'api.bob.vaspbot.net', '14WU745djqecaJ1gmtWQGeMCFim1W5MNp3');


INSERT INTO vasps(vasp_id, display_name, description, private_key, public_key, websocket_address, trisa_ds_id, trisa_ds_name, trisa_protocol_host)
 VALUES('api.alice.vaspbot.net', 'Alice VASP', 'A small VASP.  Registered in the Trisa Directory Service.', 'private', 'MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCXxiMdmxtfm3zIU2Fv8c8ctfx1U7r0vmNRsYmlBsC+KUzuu4KLSMFPQooz7zaTP5SmRrXDu2dx29EU3YtndcQRxGOdAJ06uiprEwAidHEAS+dCm+Cm+4iZLwJwG/AuzTdEz5zJXlTsZS5NgXAbBJ/tjPhIvLXNDaa3ZaDXCzOf9QIDAQAB', 'admin.alice.vaspbot.net:443', '2', 'Alices VASP', 'http://46.124.45.3:4444');

INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('mary@alicevasp.us', 'api.alice.vaspbot.net', '1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('alice@alicevasp.us', 'api.alice.vaspbot.net', '1MRCxvEpBoY8qajrmNTSrcfXSZ2wsrGeha');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('jane@alicevasp.us', 'api.alice.vaspbot.net', '14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v');



INSERT INTO vasps(vasp_id, display_name, description, private_key, public_key, websocket_address, trisa_ds_id, trisa_ds_name, trisa_protocol_host)
 VALUES('EVIL-GUID', 'Evil VASP', 'An evil VASP out to do no good.  NOT registered in Trisa Directory Service', 'private', 'N/A', '127.0.0.1:4437', null, null, null);

INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('voldemort', 'EVIL-GUID', 'mary@evilvasp');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('launderer', 'EVIL-GUID', 'alice@evilvasp');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('badnews', 'EVIL-GUID', 'jane@evilvasp');

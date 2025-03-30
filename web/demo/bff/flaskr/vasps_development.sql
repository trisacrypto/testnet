INSERT INTO vasps(vasp_id, display_name, description, private_key, public_key, websocket_address, trisa_ds_id, trisa_ds_name, trisa_protocol_host)
 VALUES('api.bob.vaspbot.com', 'Bob VASP', 'Run of the mill VASP.  Registered in the Trisa Directory Service.', 'private', 'E9Lmo/wLSj8Q0FqFhxiTl7r+y2U4U78YjJ/c3sCo20+hwkHDzn1WjII9TLk/dpJUIwaAx5pB/nGeOWeNLzBTZZspXMWAJSORdZyWtOe6jKBsRAfIW/qKeqltg5O3GkiZzC74toZwuNU7W1IkWzIiT6l7qlWlRjPTo4gYmtRSB/W4ZYBiCbAW9YnOUw0Qt2VYKo3N02EeuZAYBoem57RUgOwgR4+sBx3K67JRiR1JVdm9JQ5FUY4QRCQoMbXV0gVCipET/lMLxFJVQ0eTxdmqiV0CwBLege1mDY0CCdBUadizQC3IjGAr8kvcbZa/DEX3nS5j1sMsA50ZOQivjom1SUCyPy6Nj/05RHbVDXDIh9sNSqTJ0qpK1aFpjuoQu0DTKESGs/LoL6yCq3N/LhMetawZjSCsRaMeppe45ZD7sSosOCmATIymgY02TexPZdVNTZRr87gPk8u+vplThzKu4HSbmkPfvBwQ2PZcJIaMq/DU3RmkveYoqdiWoi/90tTPMpGdfwj5N+6TJgDlp9oROImyAGy+EK4kzO3+aqvmDV9bvyewZCdsGjjqmG9jLLMOsmzqNzDMu/M/Hs5jqECB9ByARQ2S7tNVZpt6nl/wDiFzMAqnj54dNFbPlrkOan84Ie9/p7ZaJbu91r5DipCAo7LO0M2slCygsvBJdpdgS2E=', 'admin.bob.vaspbot.com:443', '9e069e01-8515-4d57-b9a5-e249f7ab4fca', 'Bobs VASP', 'http://api.bob.vaspbot.com:443');

INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('robert@bobvasp.co.uk', 'api.bob.vaspbot.com', '18nxAxBktHZDrMoJ3N2fk9imLX8xNnYbNh');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('george@bobvasp.co.uk', 'api.bob.vaspbot.com', '1LgtLYkpaXhHDu1Ngh7x9fcBs5KuThbSzw');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('larry@bobvasp.co.uk', 'api.bob.vaspbot.com', '14WU745djqecaJ1gmtWQGeMCFim1W5MNp3');


INSERT INTO vasps(vasp_id, display_name, description, private_key, public_key, websocket_address, trisa_ds_id, trisa_ds_name, trisa_protocol_host)
 VALUES('api.alice.vaspbot.com', 'AliceCoin', 'A small VASP.  Registered in the Trisa Directory Service.', 'private', 'UA5tAqUTULJfNMmLTi0kT+j4s95+iLu3t+4ambDoKBlhNziYII1WtnEAh5rOlr6/VyjGcR604qN8SWeK4q3j3kvi0t1ONNUaiQRW+uwe4mjp03ywdLCfyEysepZLTPM6G6ykZD+dyL3Yf57VPsx2sWrC5hc/qqAb6D+yVH6LOKKn9p8JclfygxfXya7QHUJhd7oUroNI1GoPWomu8hKO0JM0VXAb7q7xjVd8npL9iT41LCueJD2+A07Yvm0dQI1Qc4UgS4/TQp/DTNlSqa7cr6EF8ZsRWRnJOdgVPOWJ9+s74OLfWNw5mIwBe3dtHKO605Dr6GQ02yrs5qkfNnHa/pYVIrG49Dmtlar9OF9J39OifIv4LHTaHNJunMqYloX5T1M3iSZIk/uGfs3k4Rt18wcttMYWp9gk/Xxqlg7N17KtE15csYW9gWhRBDxbS8juQ5Zu7+BjAWc6OdzMcNYDFbJ7XG9mpPt/9g+VxJq1xRlIKGlRnGJ+gTU/bcao8a/fHxsl1bwLOQpnO2SmciDbjLh3bmawERMk9Ac3V3GmS9wnCILmOb4s8tgP21Yd9e73aHXVgQOpJcZYSMJzN6BbhWiw7jYON7euW/lIisQr2QAdHzJSkT1sm2nEV8j9qzNhO9/gK/cWQYzoGRkR7At5nR8AMVKl962GeZ8zahPB09w=', 'admin.alice.vaspbot.com:443', '7a96ca2c-2818-4106-932e-1bcfd743b04c', 'AliceCoin', 'http://api.alice.vaspbot.com:443');

INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('mary@alicevasp.us', 'api.alice.vaspbot.com', '1ASkqdo1hvydosVRvRv2j6eNnWpWLHucMX');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('alice@alicevasp.us', 'api.alice.vaspbot.com', '1MRCxvEpBoY8qajrmNTSrcfXSZ2wsrGeha');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('jane@alicevasp.us', 'api.alice.vaspbot.com', '14HmBSwec8XrcWge9Zi1ZngNia64u3Wd2v');



INSERT INTO vasps(vasp_id, display_name, description, private_key, public_key, websocket_address, trisa_ds_id, trisa_ds_name, trisa_protocol_host)
 VALUES('api.evil.vaspbot.com', 'Evil VASP', 'An evil VASP out to do no good.  NOT registered in Trisa Directory Service', 'private', 'N/A', 'admin.evil.vaspbot.com:443', null, 'Evil VASP', 'http://api.evil.vaspbot.com:443');

INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('voldemort@evilvasp.gg', 'api.evil.vaspbot.com', '1PFTsUQrRqvmFkJunfuQbSC2k9p4RfxYLF');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('launderer@evilvasp.gg', 'api.evil.vaspbot.com', '172n89jLjXKmFJni1vwV5EzxKRXuAAoxUz');
INSERT INTO wallets(wallet_address, vasp_id, wallet_id) VALUES
 ('badnews@evilvasp.gg', 'api.evil.vaspbot.com', '182kF4mb5SW4KGEvBSbyXTpDWy8rK1Dpu');

#!/usr/bin/env python3

import os
import json
import base64
import random
import hashlib
import argparse
import psycopg2

from datetime import datetime

def clean(conn, tables):
    cur = conn.cursor()
    for table in tables:
        cur.execute(f"DELETE FROM {table}")
    cur.close()
    conn.commit()


def create_vasps(conn, vasp):
    params = []
    sql = "INSERT INTO vasps (id, name, legal_name, is_local, ivms101, created_at, updated_at) VALUES (?,?,?,?,?,?,?)"
    cur = conn.cursor()

    for i, (name, record) in enumerate(VASPS.items()):
        # TODO: look up VASP ID in Directory Service
        ts = datetime.now()
        is_local = name == vasp
        legal_name = record["legal_person"]["name"]["name_identifiers"][0]["legal_person_name"]
        common_name = record["common_name"]

        # Only store IVMS data if this is the local VASP
        # (so that VASPs have to look each other up in the directory service)
        record = json.dumps({"legal_person": record["legal_person"]}) if is_local else None
        params.append([i+1, common_name, legal_name, is_local, record, ts, ts])

    cur.executemany(sql, params)


def create_wallets(conn, vasp):
    params = []
    cur = conn.cursor()
    sql = "INSERT INTO wallets (address, email, provider_id, created_at, updated_at) VALUES (?,?,?,?,?)"

    for wallet in WALLETS:
        ts = datetime.now()
        params.append([wallet[0], wallet[1], wallet[2], ts, ts])

    cur.executemany(sql, params)


def create_accounts(conn, vasp):
    params = []
    cur = conn.cursor()
    sql = "INSERT INTO accounts (name, email, wallet_address, ivms101, balance, created_at, updated_at) VALUES (?,?,?,?,?,?,?)"

    for wallet in WALLETS:
        domain = wallet[1].split("@")[-1]
        if not domain.startswith(vasp):
            continue

        # If the wallet belongs to the VASP assign it a "customer identification"
        wallet[3]["natural_person"]["customer_identification"] = str(random.randint(100, 10000))

        # Get the name of the person for the account
        name_parts = wallet[3]["natural_person"]["name"]["name_identifiers"][0]
        name = "{secondary_identifier} {primary_identifier}".format(**name_parts)
        ivms = json.dumps(wallet[3])
        ts = datetime.now()

        # Give the account a random positive balance
        balance = random.randint(50, 5000) + (random.randint(0, 100) / 100)
        params.append([name, wallet[1], wallet[0], ivms, balance, ts, ts])

    cur.executemany(sql, params)


def create_transactions(conn, vasp):
    cur = conn.cursor()
    acc = "SELECT id FROM accounts WHERE email=?"
    idn = "INSERT INTO identities (wallet_address, email, provider) VALUES (?,?,?)"
    trn = "INSERT INTO transactions (account_id, originator_id, beneficiary_id, amount, debit, completed, timestamp, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?)"
    idg = "SELECT id FROM identities WHERE hash=?"

    for tx in TRANSACTIONS:
        txts = datetime.now()

        # Insert originator identity
        originator = find_wallet(tx["originator"])
        oprovider = find_provider(tx["originator"])
        ohash = identity_signature(originator[0], json.dumps(originator[3]), json.dumps(oprovider))
        cur.execute(idn, [originator[0], json.dumps(originator[3]), json.dumps(oprovider), ohash])
        cur.execute(idg, (ohash,))
        originator_id = cur.fetchone()[0]

        # Insert beneficiary identity
        beneficiary = find_wallet(tx["beneficiary"])
        bprovider = find_provider(tx["beneficiary"])
        bhash = identity_signature(beneficiary[0], json.dumps(beneficiary[3]), json.dumps(bprovider))
        cur.execute(idn, [beneficiary[0], json.dumps(beneficiary[3]), json.dumps(bprovider), bhash])
        cur.execute(idg, (bhash,))
        beneficiary_id = cur.fetchone()[0]

        if tx["originator"].split("@")[-1].startswith(vasp):
            # handle originator side transaction insert
            cur.execute(acc, (tx["originator"],))
            account_id = cur.fetchone()[0]
            cur.execute(trn, (account_id, originator_id, beneficiary_id, tx["amount"], True, True, txts, datetime.now(), datetime.now()))

        if tx["beneficiary"].split("@")[-1].startswith(vasp):
            # handle beneficiary side transaction insert
            cur.execute(acc, (tx["beneficiary"],))
            account_id = cur.fetchone()[0]
            cur.execute(trn, (account_id, originator_id, beneficiary_id, tx["amount"], False, True, txts, datetime.now(), datetime.now()))


def find_wallet(email):
    for wallet in WALLETS:
        if wallet[1] == email:
            return wallet
    raise ValueError(f"could not find wallet for {email}")


def find_provider(email):
    domain = email.split("@")[-1]
    for name, data in VASPS.items():
        if domain.startswith(name):
            return data
    raise ValueError(f"could not find provider for {email}")


def identity_signature(wallet_address, identity, provider):
    m = hashlib.sha3_256()
    m.update(wallet_address.encode("utf-8"))
    m.update(identity.encode("utf-8"))
    m.update(provider.encode("utf-8"))
    return base64.b64encode(m.digest())


def main(args):
    with psycopg2.connect(args.db) as conn:
        if args.clean:
            clean(conn)

        create_vasps(conn, args.vasp)
        create_wallets(conn, args.vasp)
        create_accounts(conn, args.vasp)
        create_transactions(conn, args.vasp)
        conn.commit()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="generates database fixtures for an rVASP"
    )
    parser.add_argument(
        "-v", "--vasp",
        choices={"bob", "alice", "evil"},
        default=os.getenv("RVASP_NAME"),
        help="name of the VASP to generate the database for",
    )
    parser.add_argument(
        "-c", "--clean", action="store_true",
        help="clean up anything in the tables before populating",
    )
    parser.add_argument(
        "-d", "--db",
        default=os.getenv("RVASP_DATABASE", "rvasp.db"),
        help="DSN to postgres database to connect to",
    )

    args = parser.parse_args()
    main(args)
import os
import json
import shutil
import argparse

VASPS_FILE = 'vasps.json'
WALLETS_FILE = 'wallets.json'

def migrate_fixtures(args):
    # Load the VASPs fixtures
    with open(os.path.join(args.fixtures, VASPS_FILE), 'r') as rvasp_file:
        rvasps = json.load(rvasp_file)

    names = args.names.split(',')

    # Insert the rVASP name into the fixture
    for record in rvasps:
        for n in names:
            if n in record['common_name']:
                record['common_name'] = n
                break

    # Write the rVASP fixtures to the output path
    with open(os.path.join(args.output, VASPS_FILE), 'w') as output_file:
        json.dump(rvasps, output_file, indent=4)

    # Copy the wallets fixtures to the output path
    shutil.copy(os.path.join(args.fixtures, WALLETS_FILE), os.path.join(args.output, WALLETS_FILE))

def main():
    parser = argparse.ArgumentParser(description='rVASP Fixture Migrator')
    parser.add_argument('-f', '--fixtures', help='Path to the rVASP fixtures', required=True)
    parser.add_argument('-o', '--output', help='Path to the output directory for the fixtures', required=True)
    parser.add_argument('-n', '--names', help='Comma-separated list of VASP names', required=True)

    args = parser.parse_args()
    migrate_fixtures(args)

    print('rVASP fixtures migrated to {}'.format(args.output))

if __name__ == '__main__':
    main()
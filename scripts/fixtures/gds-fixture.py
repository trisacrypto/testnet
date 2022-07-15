import pem
import json
import base64
import argparse

def generate_fixtures(args):
    # Load the GDS template JSON
    with open(args.template, 'r') as template_file:
        template = json.load(template_file)
    
    # Insert the name, id, and TRISA endpoint into the template
    template['common_name'] = args.name
    template['entity']['name']['name_identifiers'][0]['legal_person_name'] = args.name
    template['id'] = args.id
    template['trisa_endpoint'] = args.endpoint

    # Parse the pem certificate into a base64 string
    certs = pem.parse_file(args.cert)
    encoded = base64.b64encode(certs[0].as_bytes())
    template['signing_certificates'] = [{'data': encoded.decode('ascii')}]

    # Write the GDS fixture to the output path
    with open(args.output, 'w') as output_file:
        json.dump(template, output_file, indent=4)

def main():
    parser = argparse.ArgumentParser(description='GDS VASP Fixture Generator')
    parser.add_argument('-t', '--template', help='Path to template JSON to base fixture on', required=True)
    parser.add_argument('-o', '--output', help='Output path for the GDS VASP fixture', required=True)
    parser.add_argument('-n', '--name', help='Common name of the VASP', required=True)
    parser.add_argument('-i', '--id', help='ID for the VASP', required=True)
    parser.add_argument('-c', '--cert', help='Path to the public certificate to store in the VASP', required=True)
    parser.add_argument('-e', '--endpoint', help='TRISA endpoint for the VASP', required=True)

    args = parser.parse_args()
    generate_fixtures(args)

    print('GDS VASP fixture generated at {}'.format(args.output))

if __name__ == '__main__':
    main()
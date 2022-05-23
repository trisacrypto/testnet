import json
import argparse

def generate_fixtures(template_path, output_path, name, id, endpoint):
    # Load the template JSON
    with open(template_path, 'r') as template_file:
        template = json.load(template_file)
    
    # Insert the name, id, and TRISA endpoint into the template
    template['common_name'] = name
    template['entity']['name']['name_identifiers'][0]['legal_person_name'] = name
    template['id'] = id
    template['trisa_endpoint'] = endpoint

    # Write the fixture to the output path
    with open(output_path, 'w') as output_file:
        json.dump(template, output_file, indent=4)

def main():
    parser = argparse.ArgumentParser(description='GDS VASP Fixture Generator')
    parser.add_argument('-t', '--template', help='Path to template JSON to base fixture on', required=True)
    parser.add_argument('-o', '--output', help='Output path for the GDS VASP fixture', required=True)
    parser.add_argument('-n', '--name', help='Common name of the VASP', required=True)
    parser.add_argument('-i', '--id', help='ID for the VASP', required=True)
    parser.add_argument('-e', '--endpoint', help='TRISA endpoint for the VASP', required=True)

    args = parser.parse_args()
    generate_fixtures(args.template, args.output, args.name, args.id, args.endpoint)

if __name__ == '__main__':
    main()
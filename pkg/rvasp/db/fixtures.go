package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/shopspring/decimal"
)

const (
	VASPS_FILE        = "vasps.json"
	WALLETS_FILE      = "wallets.json"
	TRANSACTIONS_FILE = "transactions.json"
)

// Load unstructured JSON data from a file in the fixtures directory. The caller must
// handle converting the data to the desired type.
func loadFile(dir, fixture string) (bytes []byte, err error) {
	var (
		file *os.File
	)

	if file, err = os.Open(filepath.Join(dir, fixture)); err != nil {
		return nil, err
	}
	defer file.Close()

	if bytes, err = ioutil.ReadAll(file); err != nil {
		return nil, err
	}

	return bytes, nil
}

// Helper function that performs successive key lookups on a map object and returns the
// final value. This assumes that all keys in the "chain" point except the final one
// point to a map object and raises an error if any key in the chain is not found.
func lookupKeys(obj map[string]interface{}, keys ...string) (value interface{}, err error) {
	value = obj
	for _, key := range keys {
		var ok bool

		if value, ok = value.(map[string]interface{})[key]; !ok {
			return nil, fmt.Errorf("key %s not found in object %v", key, obj)
		}
	}

	return value, nil
}

// Load VASPS from the fixtures directory as a slice of VASP objects.
func LoadVASPs(fixturesPath string) (vasps []VASP, err error) {
	var (
		bytes []byte
		obj   []interface{}
	)
	if bytes, err = loadFile(fixturesPath, VASPS_FILE); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &obj); err != nil {
		return nil, err
	}

	for _, record := range obj {
		v := VASP{}
		var ok bool
		var obj interface{}

		// Parse common name
		if v.Name, ok = record.(map[string]interface{})["common_name"].(string); !ok {
			return nil, fmt.Errorf("could not parse common name for vasp: %v", record)
		}

		// Parse IVMS101
		var legal_person interface{}
		if legal_person, ok = record.(map[string]interface{})["legal_person"]; !ok {
			return nil, fmt.Errorf("could not parse legal person for vasp: %v", record)
		}

		// Dump legal person to JSON string
		legal_person_json := map[string]interface{}{
			"legal_person": legal_person,
		}
		var legal_person_bytes []byte
		if legal_person_bytes, err = json.Marshal(legal_person_json); err != nil {
			return nil, err
		}
		v.IVMS101 = string(legal_person_bytes)

		// Parse legal name
		if obj, err = lookupKeys(legal_person.(map[string]interface{}), "name", "name_identifiers"); err != nil {
			return nil, fmt.Errorf("could not parse legal ids for vasp: %s", err)
		}

		if len(obj.([]interface{})) == 0 {
			return nil, fmt.Errorf("no legal ids found for vasp: %s", err)
		}

		var legal_id map[string]interface{}
		if legal_id, ok = obj.([]interface{})[0].(map[string]interface{}); !ok {
			return nil, fmt.Errorf("could not parse legal id for vasp: %v", obj)
		}

		var legal_name string
		if legal_name, ok = legal_id["legal_person_name"].(string); !ok {
			return nil, fmt.Errorf("could not parse legal name for vasp: %v", legal_id)
		}
		v.LegalName = &legal_name

		vasps = append(vasps, v)
	}

	return vasps, nil
}

// Load wallets from the fixtures directory as a slice of Wallet objects. Also returns
// a slice of Account objects which are derived from the wallet values.
func LoadWallets(fixturesPath string) (wallets []Wallet, accounts []Account, err error) {
	var (
		bytes []byte
		obj   []interface{}
	)

	if bytes, err = loadFile(fixturesPath, WALLETS_FILE); err != nil {
		return nil, nil, err
	}

	if err = json.Unmarshal(bytes, &obj); err != nil {
		return nil, nil, err
	}

	for _, record := range obj {
		w := Wallet{}
		a := Account{}

		// Validate wallet record
		var fields []interface{}
		var ok bool
		if fields, ok = record.([]interface{}); !ok {
			return nil, nil, fmt.Errorf("could not parse wallet record: %v", record)
		}

		// Validate the number of fields
		if len(fields) != 5 {
			return nil, nil, fmt.Errorf("invalid number of wallet fields: got %d, expected 5", len(fields))
		}

		// Parse the wallet fields
		w.Address = fields[0].(string)
		w.Email = fields[1].(string)
		w.ProviderID = uint(fields[2].(float64))
		w.VaspID = w.ProviderID

		a.Email = w.Email
		a.WalletAddress = w.Address
		a.VaspID = w.ProviderID

		// Validate the policy field
		w.Policy = PolicyType(fields[3].(string))
		if !isValidPolicy(w.Policy) {
			return nil, nil, fmt.Errorf("invalid policy for wallet %s: %s", w.Address, w.Policy)
		}

		// Parse the account name
		var person map[string]interface{}
		if person, ok = fields[4].(map[string]interface{}); !ok {
			return nil, nil, fmt.Errorf("could not parse person record: %v", record)
		}
		var name_ids interface{}
		if name_ids, err = lookupKeys(person, "natural_person", "name", "name_identifiers"); err != nil {
			return nil, nil, fmt.Errorf("could not parse name identifiers for wallet %s: %s", w.Address, err)
		}

		if len(name_ids.([]interface{})) == 0 {
			return nil, nil, fmt.Errorf("no name identifiers found for wallet %s", w.Address)
		}

		var name_id map[string]interface{}
		if name_id, ok = name_ids.([]interface{})[0].(map[string]interface{}); !ok {
			return nil, nil, fmt.Errorf("could not parse name identifier for wallet %s: %v", w.Address, name_ids)
		}

		var primary_id, secondary_id string
		if primary_id, ok = name_id["primary_identifier"].(string); !ok {
			return nil, nil, fmt.Errorf("could not parse primary identifier for wallet %s: %v", w.Address, name_id)
		}
		if secondary_id, ok = name_id["secondary_identifier"].(string); !ok {
			return nil, nil, fmt.Errorf("could not parse secondary identifier for wallet %s: %v", w.Address, name_id)
		}

		a.Name = fmt.Sprintf("%s %s", secondary_id, primary_id)

		// Parse the account IVMS101
		var natural_person_bytes []byte
		if natural_person_bytes, err = json.Marshal(person); err != nil {
			return nil, nil, err
		}
		a.IVMS101 = string(natural_person_bytes)

		// Give the account a random positive balance
		a.Balance = decimal.NewFromFloat32(float32(rand.Intn(4950) + 50 + (rand.Intn(100) / 100.0)))

		wallets = append(wallets, w)
		accounts = append(accounts, a)
	}

	return wallets, accounts, nil
}

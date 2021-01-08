package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	pb "github.com/trisacrypto/testnet/pkg/trisads/pb/models/v1alpha1"
	"github.com/urfave/cli"
	"google.golang.org/protobuf/proto"
)

func main() {
	app := cli.NewApp()
	app.Name = "debug"
	app.Version = "alpha"
	app.Usage = "debugging utilities for the TRISA TestNet"
	app.Commands = []cli.Command{
		{
			Name:     "store:keys",
			Usage:    "list the keys currently in the leveldb store",
			Category: "store",
			Action:   storeKeys,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "dsn to connect to trisa directory storage",
					EnvVar: "TRISADS_DATABASE",
				},
				cli.BoolFlag{
					Name:  "s, stringify",
					Usage: "stringify keys otherwise they are base64 encoded",
				},
				cli.StringFlag{
					Name:  "p, prefix",
					Usage: "specify a prefix to filter keys on",
				},
			},
		},
		{
			Name:      "store:get",
			Usage:     "get the value for the specified key",
			Category:  "store",
			Action:    storeGet,
			ArgsUsage: "key [key ...]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "dsn to connect to trisa directory storage",
					EnvVar: "TRISADS_DATABASE",
				},
				cli.BoolFlag{
					Name:  "b, b64decode",
					Usage: "specify the keys as base64 encoded values which must be decoded",
				},
			},
		},
		{
			Name:     "store:put",
			Usage:    "put the value for the specified key",
			Category: "store",
			Action:   storePut,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "dsn to connect to trisa directory storage",
					EnvVar: "TRISADS_DATABASE",
				},
				cli.BoolFlag{
					Name:  "b, b64decode",
					Usage: "specify the key and value as base64 encoded strings which must be decoded",
				},
				cli.StringFlag{
					Name:  "k, key",
					Usage: "the key to put the value to",
				},
				cli.StringFlag{
					Name:  "v, value",
					Usage: "the value to put to the database (or specify json document)",
				},
				cli.StringFlag{
					Name:  "p, path",
					Usage: "path to a JSON document containing the value",
				},
			},
		},
		{
			Name:      "store:delete",
			Usage:     "delete the leveldb record for the specified key(s)",
			Category:  "store",
			Action:    storeDelete,
			ArgsUsage: "key [key ...]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "dsn to connect to trisa directory storage",
					EnvVar: "TRISADS_DATABASE",
				},
				cli.BoolFlag{
					Name:  "b, b64decode",
					Usage: "specify the keys as base64 encoded values which must be decoded",
				},
			},
		},
		{
			Name:      "decrypt",
			Usage:     "decrypt base64 encoded ciphertext with an HMAC signature",
			ArgsUsage: "ciphertext hmac",
			Category:  "cipher",
			Action:    cipherDecrypt,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "k, key",
					Usage:  "secret key to decrypt the cipher text",
					EnvVar: "TRISADS_SECRET_KEY",
				},
			},
		},
	}

	app.Run(os.Args)
}

func storeKeys(c *cli.Context) (err error) {
	if c.String("db") == "" {
		return cli.NewExitError("specify path to leveldb database", 1)
	}

	var db *leveldb.DB
	if db, err = leveldb.OpenFile(c.String("db"), nil); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer db.Close()

	var prefix *util.Range
	if prefixs := c.String("prefix"); prefixs != "" {
		prefix = util.BytesPrefix([]byte(prefixs))
	}

	iter := db.NewIterator(prefix, nil)
	defer iter.Release()

	stringify := c.Bool("stringify")
	for iter.Next() {
		if stringify {
			fmt.Printf("- %s\n", string(iter.Key()))
		} else {
			fmt.Printf("- %s\n", base64.RawStdEncoding.EncodeToString(iter.Key()))
		}
	}

	if err = iter.Error(); err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}

func storeGet(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return cli.NewExitError("specify at least one key to fetch", 1)
	}
	if c.String("db") == "" {
		return cli.NewExitError("specify path to leveldb database", 1)
	}

	var db *leveldb.DB
	if db, err = leveldb.OpenFile(c.String("db"), nil); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer db.Close()

	b64decode := c.Bool("b64decode")
	for _, keys := range c.Args() {
		var key []byte
		if b64decode {
			if key, err = base64.RawStdEncoding.DecodeString(keys); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else {
			key = []byte(keys)
		}

		var data []byte
		if data, err = db.Get(key, nil); err != nil {
			return cli.NewExitError(err, 1)
		}

		// Unmarshall the thing
		var value interface{}

		// Determine how to unmarshall the data
		if bytes.HasPrefix(key, []byte("vasps")) {
			vasp := new(pb.VASP)
			if err = proto.Unmarshal(data, vasp); err != nil {
				return cli.NewExitError(err, 1)
			}
			value = vasp
		} else if bytes.HasPrefix(key, []byte("certreqs")) {
			careq := new(pb.CertificateRequest)
			if err = proto.Unmarshal(data, careq); err != nil {
				return cli.NewExitError(err, 1)
			}
			value = careq
		} else if bytes.Equal(key, []byte("index::names")) {
			value = make(map[string]string)
			if err = json.Unmarshal(data, &value); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else if bytes.Equal(key, []byte("index::countries")) {
			value = make(map[string][]string)
			if err = json.Unmarshal(data, &value); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else if bytes.Equal(key, []byte("sequence::pks")) {
			pk, n := binary.Uvarint(data)
			if n <= 0 {
				return cli.NewExitError("could not parse sequence", 1)
			}
			value = pk
		} else {
			return cli.NewExitError("could not determine unmarshall type", 1)
		}

		// Marshall the JSON representation
		var out []byte
		if out, err = json.MarshalIndent(value, "", "  "); err != nil {
			return cli.NewExitError(err, 1)
		}
		fmt.Println(string(out))
	}

	return nil
}

func storePut(c *cli.Context) (err error) {
	if c.String("key") == "" {
		return cli.NewExitError("must specify a key to put to", 1)
	}
	if c.String("value") != "" && c.String("path") != "" {
		return cli.NewExitError("specify either value or path, not both", 1)
	}
	if c.String("db") == "" {
		return cli.NewExitError("specify path to leveldb database", 1)
	}

	var db *leveldb.DB
	if db, err = leveldb.OpenFile(c.String("db"), nil); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer db.Close()

	var key, data, value []byte
	keys := c.String("key")
	b64decode := c.Bool("b64decode")

	if b64decode {
		if key, err = base64.RawStdEncoding.DecodeString(keys); err != nil {
			return cli.NewExitError(err, 1)
		}
	} else {
		key = []byte(keys)
	}

	if c.String("value") != "" {
		if b64decode {
			// If value is b64 encoded then we just assume it's data to put directly
			if value, err = base64.RawStdEncoding.DecodeString(c.String("value")); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else {
			data = []byte(keys)
		}
	}

	if c.String("path") != "" {
		if data, err = ioutil.ReadFile(c.String("path")); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	// Quick spot check
	if len(data) == 0 && len(value) == 0 {
		return cli.NewExitError("no value to put to database", 1)
	}

	if len(data) > 0 && len(value) > 0 {
		return cli.NewExitError("both data and value specified?", 1)
	}

	if len(data) > 0 {
		// Unmarshall the thing from JSON then
		// Marshall the database representation
		if bytes.HasPrefix(key, []byte("vasps")) {
			vasp := new(pb.VASP)
			if err = json.Unmarshal(data, &vasp); err != nil {
				return cli.NewExitError(err, 1)
			}
			if value, err = proto.Marshal(vasp); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else if bytes.HasPrefix(key, []byte("certreqs")) {
			careq := new(pb.CertificateRequest)
			if err = json.Unmarshal(data, &careq); err != nil {
				return cli.NewExitError(err, 1)
			}
			if value, err = proto.Marshal(careq); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else if bytes.Equal(key, []byte("index::names")) {
			var names map[string]string
			if err = json.Unmarshal(data, &names); err != nil {
				return cli.NewExitError(err, 1)
			}
			if value, err = json.Marshal(names); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else if bytes.Equal(key, []byte("index::countries")) {
			var countries map[string][]string
			if err = json.Unmarshal(data, &countries); err != nil {
				return cli.NewExitError(err, 1)
			}
			if value, err = json.Marshal(countries); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else if bytes.Equal(key, []byte("sequence::pks")) {
			var pk uint64
			if err = json.Unmarshal(data, &pk); err != nil {
				return cli.NewExitError(err, 1)
			}
			value = make([]byte, binary.MaxVarintLen64)
			binary.PutUvarint(value, pk)
		} else {
			return cli.NewExitError("could not determine unmarshall type", 1)
		}
	}

	// Final spot check
	if len(value) == 0 {
		return cli.NewExitError("no value marshalled", 1)
	}

	// Put the key/value to the database
	if err = db.Put(key, value, nil); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

func storeDelete(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return cli.NewExitError("specify at least one key to fetch", 1)
	}
	if c.String("db") == "" {
		return cli.NewExitError("specify path to leveldb database", 1)
	}

	var db *leveldb.DB
	if db, err = leveldb.OpenFile(c.String("db"), nil); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer db.Close()

	b64decode := c.Bool("b64decode")
	for _, keys := range c.Args() {
		var key []byte
		if b64decode {
			if key, err = base64.RawStdEncoding.DecodeString(keys); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else {
			key = []byte(keys)
		}

		if err = db.Delete(key, nil); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	return nil
}

// TODO: package this all up somewhere!

const nonceSize = 12

func cipherDecrypt(c *cli.Context) (err error) {
	if c.NArg() != 2 {
		return cli.NewExitError("must specify ciphertext and hmac arguments", 1)
	}

	var secret string
	if secret = c.String("key"); secret == "" {
		return cli.NewExitError("cipher key required", 1)
	}

	var ciphertext, signature []byte
	if ciphertext, err = base64.RawStdEncoding.DecodeString(c.Args()[0]); err != nil {
		return cli.NewExitError(fmt.Errorf("could not decode ciphertext: %s", err), 1)
	}
	if signature, err = base64.RawStdEncoding.DecodeString(c.Args()[1]); err != nil {
		return cli.NewExitError(fmt.Errorf("could not decode signature: %s", err), 1)
	}

	if len(ciphertext) == 0 {
		return cli.NewExitError("empty cipher text", 1)
	}

	// Create a 32 byte signature of the key
	hash := sha256.New()
	hash.Write([]byte(secret))
	key := hash.Sum(nil)

	// Separate the data from the nonce
	data := ciphertext[:len(ciphertext)-nonceSize]
	nonce := ciphertext[len(ciphertext)-nonceSize:]

	// Validate HMAC signature
	if err = validateHMAC(key, data, signature); err != nil {
		return cli.NewExitError(err, 1)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	plainbytes, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	fmt.Println(string(plainbytes))
	return nil
}

func createHMAC(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot sign empty data")
	}
	hm := hmac.New(sha256.New, key)
	hm.Write(data)
	return hm.Sum(nil), nil
}

func validateHMAC(key, data, sig []byte) error {
	hmac, err := createHMAC(key, data)
	if err != nil {
		return err
	}

	if !bytes.Equal(sig, hmac) {
		return errors.New("HMAC mismatch")
	}
	return nil
}

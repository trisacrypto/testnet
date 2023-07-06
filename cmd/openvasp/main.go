package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/fiatjaf/go-lnurl"
	"github.com/joho/godotenv"
	"github.com/trisacrypto/testnet/pkg"
	openvasp "github.com/trisacrypto/testnet/pkg/openvasp/web-service-gin"
	"github.com/urfave/cli"
)

func main() {
	// Load the dotenv file if it exists
	godotenv.Load()

	app := cli.NewApp()

	app.Name = "openvasp"
	app.Version = pkg.Version()
	app.Usage = "a gRPC based directory service for TRISA identity lookups"
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run the openvasp gin server",
			Category: "server",
			Action:   serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "a, address",
					Usage:  "the address and port to bind the server on",
					Value:  "localhost:4435",
					EnvVar: "OPENVASP_BIND_ADDR",
				},
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the dsn of the postgres database to connect to",
					Value:  "localhost:4434",
					EnvVar: "OPENVASP_DATABASE_DSN",
				},
			},
		},
		{
			Name:     "register",
			Usage:    "",
			Category: "client",
			Action:   register,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
				cli.StringFlag{
					Name:  "n, name",
					Usage: "name of the OpenVASP customer",
					Value: "Tildred Milcot",
				},
				cli.IntFlag{
					Name:  "t, assetType",
					Usage: "AssetType (for example bitcoin) of the OpenVASP customer",
					Value: 3,
				},
				cli.StringFlag{
					Name:  "w, walletAddress",
					Usage: "WalletAddress of the OpenVASP customer",
					Value: "926ca69a-6c22-42e6-9105-11ab5de1237b",
				},
			},
		},
		{
			Name:     "listusers",
			Usage:    "",
			Category: "client",
			Action:   listUsers,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
			},
		},
		{
			Name:     "gettraveladdress",
			Usage:    "",
			Category: "client",
			Action:   getTravelAddress,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
				cli.StringFlag{
					Name:  "i, id",
					Usage: "address of the gin server",
					Value: "b02245ba-de1e-44ed-b51b-2e93dbca426d",
				},
			},
		},
		{
			Name:     "transfer",
			Usage:    "",
			Category: "client",
			Action:   transfer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "l, lnurl",
					Usage: "lnurl encoding the address of the gin server",
					Value: "LNURL1DP68GUP69UHKCMMRV9KXSMMNWSARGDPNX5HHGUNPDEEKVETJ9UUNYDNRVYMRJCFDXE3NYV3DXSEX2D3D8YCNQDFDXYCKZC34V3JNZV3NXA3R7ARPVU7HGUNPWEJKC5N4D3J5JMN3W45HY7GQ9NV7V",
				},
				cli.StringFlag{
					Name:  "p, path",
					Usage: "path to the IVMS101 payload",
					Value: "pkg/openvasp/testdata/identity.json",
				},
				cli.StringFlag{
					Name:  "t, assetType",
					Usage: "asset type for transfer, i.e. Bitcoin, Etheriem, etc.",
					Value: "BTC",
				},
				cli.Float64Flag{
					Name:  "c, amount",
					Usage: "amount of the asset type to be transfered",
					Value: 100,
				},
			},
		},
		{
			Name:     "gettransfer",
			Usage:    "",
			Category: "client",
			Action:   getTransfer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
				cli.StringFlag{
					Name:  "i, id",
					Usage: "transfer id of the transfer to lookup",
					Value: "a6f1c411-5cc0-4867-b0eb-5f4806c70803",
				},
			},
		},
		{
			Name:     "resolve",
			Usage:    "",
			Category: "client",
			Action:   resolve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
				cli.StringFlag{
					Name:  "i, id",
					Usage: "address of the gin server",
					Value: "foo",
				},
				cli.BoolFlag{
					Name:  "y, approve",
					Usage: "asset type for transfer, i.e. Bitcoin, Etheriem, etc.",
				},
				cli.StringFlag{
					Name:  "a, payment",
					Usage: "amount of the asset type to be transfered",
					Value: "some payment address",
				},
				cli.StringFlag{
					Name:  "c, callback",
					Usage: "amount of the asset type to be transfered",
					Value: "foo",
				},
			},
		},
		{
			Name:     "confirm",
			Usage:    "",
			Category: "client",
			Action:   confirm,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
				cli.StringFlag{
					Name:  "i, id",
					Usage: "address of the gin server",
					Value: "foo",
				},
				cli.BoolFlag{
					Name:  "c, cancelled",
					Usage: "amount of the asset type to be transfered",
				},
			},
		},
	}
	app.Run(os.Args)
}

// Serve the OpenVASP gin server
func serve(c *cli.Context) (err error) {
	if err = openvasp.Serve(c.String("address"), c.String("dns")); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

//
func register(c *cli.Context) (err error) {
	url := fmt.Sprintf("http://%s/register", c.String("address"))
	body := `{"name": "Mildred Tilcott", "assettype": 3, "walletaddress": "926ca69a-6c22-42e6-9105-11ab5de1237b"}`
	var response string
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func listUsers(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/listusers", c.String("address"))
	if response, err = getRequest(url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func getTravelAddress(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/gettraveladdress/%s", c.String("address"), c.String("id"))
	if response, err = getRequest(url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func transfer(c *cli.Context) (err error) {
	var file *os.File
	if file, err = os.Open(c.String("path")); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer file.Close()

	var jsonbytes []byte
	if jsonbytes, err = ioutil.ReadAll(file); err != nil {
		return cli.NewExitError(err, 1)
	}
	ivms101 := strings.ReplaceAll(string(jsonbytes), `"`, `*`)
	ivms101 = strings.ReplaceAll(ivms101, "\n", "+")

	var url string
	if url, err = lnurl.LNURLDecode(c.String("lnurl")); err != nil {
		return cli.NewExitError(err, 1)
	}

	var response string
	body := fmt.Sprintf(`{"ivms101": "%s", "assettype": 3, "amount": 3, "callback": "foo"}`, ivms101)
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func getTransfer(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/gettransfer/%s", c.String("address"), c.String("id"))
	if response, err = getRequest(url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func resolve(c *cli.Context) (err error) {
	var body string
	if c.Bool("approve") {
		body = fmt.Sprintln(`{"approved": {"address": "some payment address", "callback: "foo"}`)
	} else {
		body = fmt.Sprintln(`{"rejected": "transfer rejected"}`)
	}

	var response string
	url := fmt.Sprintf("http://%s/inquiryresolution/%s", c.String("address"), c.String("id"))
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func confirm(c *cli.Context) (err error) {
	var body string
	if !c.Bool("cancelled") {
		body = fmt.Sprintln(`{"txid": "some asset-specific tx identifier"}`)
	} else {
		body = fmt.Sprintln(`{"canceled": "transfer canceled"}`)
	}

	var response string
	url := fmt.Sprintf("http://%s/transferconfirmation/%s", c.String("address"), c.String("id"))
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

//
func postRequest(body string, url string) (_ string, err error) {
	var request *http.Request
	byteBody := []byte(body)
	if request, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(byteBody)); err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	var response *http.Response
	if response, err = http.DefaultClient.Do(request); err != nil {
		return "", cli.NewExitError(err, 1)
	}

	var responseBody []byte
	if responseBody, err = ioutil.ReadAll(response.Body); err != nil {
		return "", cli.NewExitError(err, 1)
	}
	return string(responseBody), nil
}

//
func getRequest(url string) (_ string, err error) {
	var request *http.Request
	if request, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return "", err
	}

	var response *http.Response
	if response, err = http.DefaultClient.Do(request); err != nil {
		return "", cli.NewExitError(err, 1)
	}

	var responseBody []byte
	if responseBody, err = ioutil.ReadAll(response.Body); err != nil {
		return "", cli.NewExitError(err, 1)
	}
	return string(responseBody), nil
}

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
	app.Usage = "a tool used to run and test a Gin server implementing the TRP protocol"
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run the OpenVASP gin server",
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
			Usage:    "register a new contact with the OpenVASP server",
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
					Name:  "t, assettype",
					Usage: "asset type (for example bitcoin) of the OpenVASP customer",
					Value: 3,
				},
				cli.StringFlag{
					Name:  "w, walletaddress",
					Usage: "Wallet address of the OpenVASP customer",
					Value: "926ca69a-6c22-42e6-9105-11ab5de1237b",
				},
			},
		},
		{
			Name:     "listusers",
			Usage:    "list the registered OpenVASP contacts",
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
			Usage:    "list a specific OpenVASP contact",
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
					Usage: "The customer id of the registered user to lookup",
					Value: "b02245ba-de1e-44ed-b51b-2e93dbca426d",
				},
			},
		},
		{
			Name:     "transfer",
			Usage:    "initiate a TRP tranfer",
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
				cli.IntFlag{
					Name:  "t, assettype",
					Usage: "asset type for transfer, i.e. Bitcoin, Etheriem, etc.",
					Value: 3,
				},
				cli.Float64Flag{
					Name:  "m, amount",
					Usage: "amount of the asset type to be transfered",
					Value: 100,
				},
				cli.StringFlag{
					Name:  "c, callback",
					Usage: "callback for the beneficiary to reply to",
					Value: "foo",
				},
				cli.BoolFlag{
					Name:  "r, reject",
					Usage: "whether or not the transfer will be rejected",
				},
			},
		},
		{
			Name:     "gettransfer",
			Usage:    "list a transfer by transfer id",
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
			Usage:    "resolve a TRP transfer",
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
					Usage: "id of the transfer being confirmed",
					Value: "a6f1c411-5cc0-4867-b0eb-5f4806c70803",
				},
				cli.BoolFlag{
					Name:  "y, approve",
					Usage: "whether or not the transfer should be rejected",
				},
				cli.StringFlag{
					Name:  "a, address",
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
			Usage:    "confirms a TRP transfer after resolution",
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
					Usage: "id of the transfer being confirmed",
					Value: "a6f1c411-5cc0-4867-b0eb-5f4806c70803",
				},
				cli.BoolFlag{
					Name:  "c, cancelled",
					Usage: "whether or not the tranfer should be rejected",
				},
				cli.StringFlag{
					Name:  "x, txid",
					Usage: "the txid to be returned on approval",
					Value: "some asset-specific tx identifier",
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

// sends a POST request to the register endpoint
func register(c *cli.Context) (err error) {
	url := fmt.Sprintf("http://%s/register", c.String("address"))
	body := fmt.Sprintf(`{"name": "%s", "assettype": %d, "walletaddress": "%s"}`,
		c.String("name"),
		c.Int("assettype"),
		c.String("walletaddress"))
	var response string
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

// sends a GET request to the listusers endpoint
func listUsers(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/listusers", c.String("address"))
	if response, err = getRequest(url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

// sends a GET request to the gettraveladdress endpoint
func getTravelAddress(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/gettraveladdress/%s",
		c.String("address"),
		c.String("id"))
	if response, err = getRequest(url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

// sends a POST request to the transfer endpoint
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
	//TODO: Find a better way to avoid binding issues with quotes
	ivms101 := strings.ReplaceAll(string(jsonbytes), `"`, `*`)
	ivms101 = strings.ReplaceAll(ivms101, "\n", "+")

	var url string
	if url, err = lnurl.LNURLDecode(c.String("lnurl")); err != nil {
		return cli.NewExitError(err, 1)
	}

	var response string
	body := fmt.Sprintf(`{"ivms101": "%s", "assettype": %d, "amount": %f, "callback": "%s", "reject": "%t"}`,
		ivms101, c.Int("assettype"),
		c.Float64("amount"),
		c.String("callback"),
		c.Bool("reject"))
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

// sends a GET request to the transfer endpoint
func getTransfer(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/gettransfer/%s",
		c.String("address"),
		c.String("id"))
	if response, err = getRequest(url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

// sends a POST request to the inquiryresolution endpoint
func resolve(c *cli.Context) (err error) {
	var body string
	if c.Bool("approve") {
		body = fmt.Sprintf(`{"approved": {"address": "%s", "callback: "%s"}`,
			c.String("address"),
			c.String("callback"))
	} else {
		body = fmt.Sprintln(`{"rejected": "transfer rejected"}`)
	}

	var response string
	url := fmt.Sprintf("http://%s/inquiryresolution/%s",
		c.String("address"),
		c.String("id"))
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
		body = fmt.Sprintf(`{"txid": "%s"}`, c.String("txid"))
	} else {
		body = fmt.Sprintln(`{"canceled": "transfer canceled"}`)
	}

	var response string
	url := fmt.Sprintf("http://%s/transferconfirmation/%s",
		c.String("address"),
		c.String("id"))
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(response)
	return nil
}

// sends a POST request containing the provided body to the provided
// URL and returns the response
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

// sends a GET request to the provided URL and returns
// the response
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

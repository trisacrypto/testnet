package main

import (
	"bytes"
	"fmt"
	"io"
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
					Name:   "c, callback",
					Usage:  "the URL that the server will listen on for callbacks from the counterparty",
					Value:  "http://localhost:4435",
					EnvVar: "OPENVASP_CALLBACK_URL",
				},
				cli.StringFlag{
					Name:   "d, dsn",
					Usage:  "the dsn of the postgres database to connect to",
					Value:  "",
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
					Name:  "t, asset",
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
					Name:     "i, id",
					Usage:    "The customer id of the registered user to lookup",
					Required: true,
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
					Value: "lnurl1dp68gup69uhkcmmrv9kxsmmnwsargdpnx5hhgunpdeekvetj9uunydnrvymrjcfdxe3nyv3dxsex2d3d8ycnqdfdxyckzc34v3jnzv3nxa3r7arpvu7hgunpwejkcun4d3jkjmn3w45hy7gkp969c",
				},
				cli.BoolFlag{
					Name:  "b, beneficiary",
					Usage: "path to the IVMS101 payload",
				},
				cli.StringFlag{
					Name:  "t, asset",
					Usage: "asset type for transfer, i.e. Bitcoin, Etheriem, etc.",
					Value: "BTC",
				},
				cli.IntFlag{
					Name:  "m, amount",
					Usage: "amount of the asset type to be transfered",
					Value: 100,
				},
				cli.StringFlag{
					Name:  "c, callback",
					Usage: "callback for the beneficiary to reply to",
					Value: "foo",
				},
			},
		},
		{
			Name:     "listtransfers",
			Usage:    "list the registered OpenVASP contacts",
			Category: "client",
			Action:   listTransfers,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, address",
					Usage: "address of the gin server",
					Value: "localhost:4435",
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
					Name:     "i, id",
					Usage:    "transfer id of the transfer to lookup",
					Required: true,
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
					Name:  "e, endpoint",
					Usage: "address of the gin server",
					Value: "localhost:4435",
				},
				cli.BoolFlag{
					Name:  "r, reject",
					Usage: "whether or not the tranfer should be rejected",
				},
				cli.StringFlag{
					Name:  "c, callback",
					Usage: "amount of the asset type to be transfered",
					Value: "foo",
				},
			},
		},
	}
	app.Run(os.Args)
}

// Serve the OpenVASP gin server
func serve(c *cli.Context) (err error) {
	if err = openvasp.Serve(c.String("address"), c.String("callback"), c.String("dsn")); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

// sends a POST request to the register endpoint
func register(c *cli.Context) (err error) {
	url := fmt.Sprintf("http://%s/register", c.String("address"))
	body := fmt.Sprintf(`{"name": "%s", "assettype": %s, "walletaddress": "%s"}`,
		c.String("name"),
		c.String("asset"),
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
	var path string
	path = "pkg/openvasp/testdata/originator.json"
	if c.Bool("beneficiary") {
		path = "pkg/openvasp/testdata/beneficiary.json"
	}

	var file *os.File
	if file, err = os.Open(path); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer file.Close()

	responseBody := &bytes.Buffer{}
	if _, err = io.Copy(responseBody, file); err != nil {
		return cli.NewExitError(err, 1)
	}
	ivms101 := strings.ReplaceAll(responseBody.String(), `"`, `*`)
	ivms101 = strings.ReplaceAll(ivms101, "\n", "+")

	var url string
	if url, err = lnurl.LNURLDecode(c.String("lnurl")); err != nil {
		return cli.NewExitError(err, 1)
	}

	var response string
	body := fmt.Sprintf(`{"asset": {"slip0044": "%s"}, "amount": %d, "callback": "%s", "IVMS101": "%s"}`,
		c.String("asset"),
		c.Int("amount"),
		c.String("callback"),
		ivms101)
	if response, err = postRequest(body, url); err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Printf("\n%s\n\n", response)
	return nil
}

// sends a GET request to the listusers endpoint
func listTransfers(c *cli.Context) (err error) {
	var response string
	url := fmt.Sprintf("http://%s/listtransfers", c.String("address"))
	if response, err = getRequest(url); err != nil {
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

func confirm(c *cli.Context) (err error) {
	var body string
	if !c.Bool("reject") {
		body = fmt.Sprintf(`{"approved": {"address": "payment address", "callback": "%s"}}`, c.String("callback"))
	} else {
		body = fmt.Sprintf(`{"rejected": "transfer canceled", "callback": "%s"}`, c.String("callback"))
	}

	var response string
	if response, err = postRequest(body, c.String("endpoint")); err != nil {
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
	request.Header.Set("api-version", "3.0.0")
	request.Header.Set("request-identifier", "ebxe")

	var response *http.Response
	if response, err = http.DefaultClient.Do(request); err != nil {
		return "", cli.NewExitError(err, 1)
	}

	fmt.Println(response)
	responseBody := &bytes.Buffer{}
	if _, err = io.Copy(responseBody, response.Body); err != nil {
		return "", cli.NewExitError(err, 1)
	}
	return responseBody.String(), nil
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

	responseBody := &bytes.Buffer{}
	if _, err = io.Copy(responseBody, response.Body); err != nil {
		return "", cli.NewExitError(err, 1)
	}
	return responseBody.String(), nil
}

package main

import (
	"os"

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
					Name:  "a, addr",
					Usage: "the address and port to bind the server on",
					Value: "localhost:4435",
				},
				cli.StringFlag{
					Name:  "d, db",
					Usage: "the dsn of the postgres database to connect to",
					Value: "localhost:4434",
				},
			},
		},
	}

	app.Run(os.Args)
}

// Serve the OpenVASP gin server
func serve(c *cli.Context) (err error) {
	if err = openvasp.Serve(c.String("addr"), c.String("dns")); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

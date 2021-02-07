package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/trisacrypto/testnet/pkg"
	"github.com/trisacrypto/testnet/pkg/rvasp"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	app := cli.NewApp()

	app.Name = "rvasp"
	app.Version = pkg.Version()
	app.Usage = "a gRPC based directory service for TRISA identity lookups"
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run the rVASP service",
			Category: "server",
			Action:   serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "n, name",
					Usage: "the name of the rVASP (alice, bob, evil, etc.)",
				},
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "the address and port to bind the server on",
					Value: ":4434",
				},
				cli.StringFlag{
					Name:  "t, trisa-addr",
					Usage: "the address and port to bind the TRISA server on",
					Value: ":4435",
				},
				cli.StringFlag{
					Name:  "d, db",
					Usage: "the dsn to the sqlite3 database to connect to",
				},
			},
		},
		{
			Name:     "initdb",
			Usage:    "run the database migration",
			Category: "server",
			Action:   initdb,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the dsn to the sqlite3 database to connect to",
					Value:  "fixtures/rvasp/rvasp.db",
					EnvVar: "DATABASE_URL",
				},
			},
		},
		{
			Name:     "account",
			Usage:    "get the account status and current transactions",
			Category: "client",
			Action:   account,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "a, addr",
					Usage:  "the address and port to connect to the server on",
					Value:  "localhost:4434",
					EnvVar: "RVASP_ADDR",
				},
				cli.StringFlag{
					Name:   "e, account",
					Usage:  "the email address of the account",
					EnvVar: "RVASP_CLIENT_ACCOUNT",
				},
				cli.BoolFlag{
					Name:  "T, no-transactions",
					Usage: "don't include any transactions in the response",
				},
			},
		},
		{
			Name:     "transfer",
			Usage:    "transfer funds, initiating the TRISA protocol",
			Category: "client",
			Action:   transfer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "a, addr",
					Usage:  "the address and port to connect to the server on",
					Value:  "localhost:4434",
					EnvVar: "RVASP_ADDR",
				},
				cli.StringFlag{
					Name:   "e, account",
					Usage:  "the email address of the account",
					EnvVar: "RVASP_CLIENT_ACCOUNT",
				},
				cli.StringFlag{
					Name:  "b, beneficiary",
					Usage: "the email or wallet address of the beneficiary",
				},
				cli.Float64Flag{
					Name:  "d, amount",
					Usage: "the amount to transfer to the beneficiary",
				},
			},
		},
	}

	app.Run(os.Args)
}

// Serve the TRISA directory service
func serve(c *cli.Context) (err error) {
	var conf *rvasp.Settings
	if conf, err = rvasp.Config(); err != nil {
		return cli.NewExitError(err, 1)
	}

	if name := c.String("name"); name != "" {
		conf.Name = name
	}

	if addr := c.String("addr"); addr != "" {
		conf.BindAddr = addr
	}

	if addr := c.String("trisa-addr"); addr != "" {
		conf.TRISABindAddr = addr
	}

	if db := c.String("db"); db != "" {
		conf.DatabaseDSN = db
	}

	var srv *rvasp.Server
	if srv, err = rvasp.New(conf); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = srv.Serve(); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

// Run the database migration
func initdb(c *cli.Context) (err error) {
	var db *gorm.DB
	if db, err = gorm.Open(sqlite.Open(c.String("db")), &gorm.Config{}); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = rvasp.MigrateDB(db); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

// Client method: get account status
func account(c *cli.Context) (err error) {
	req := &pb.AccountRequest{
		Account:        c.String("account"),
		NoTransactions: c.Bool("no-transactions"),
	}

	if req.Account == "" {
		return cli.NewExitError("specify account email", 1)
	}

	client, err := makeClient(c)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rep, err := client.AccountStatus(ctx, req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return printJSON(rep)
}

// Client method: transfer funds
func transfer(c *cli.Context) (err error) {
	req := &pb.TransferRequest{
		Account:     c.String("account"),
		Beneficiary: c.String("beneficiary"),
		Amount:      float32(c.Float64("amount")),
	}

	if req.Account == "" {
		return cli.NewExitError("specify account email", 1)
	}

	if req.Beneficiary == "" {
		return cli.NewExitError("specify a beneficiary email or wallet", 1)
	}

	if req.Amount <= 0.0 {
		return cli.NewExitError("specify a transfer amount", 1)
	}

	client, err := makeClient(c)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rep, err := client.Transfer(ctx, req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return printJSON(rep)
}

// helper function to create the GRPC client with default options
func makeClient(c *cli.Context) (_ pb.TRISAIntegrationClient, err error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	var cc *grpc.ClientConn
	if cc, err = grpc.Dial(c.String("addr"), opts...); err != nil {
		return nil, err
	}
	return pb.NewTRISAIntegrationClient(cc), nil
}

// helper function to print JSON response and exit
func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	fmt.Println(string(data))
	return nil
}

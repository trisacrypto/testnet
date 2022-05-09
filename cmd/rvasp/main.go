package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/trisacrypto/testnet/pkg"
	"github.com/trisacrypto/testnet/pkg/rvasp"
	pb "github.com/trisacrypto/testnet/pkg/rvasp/pb/v1"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load the dotenv file if it exists
	godotenv.Load()

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
					Name:   "n, name",
					Usage:  "the name of the rVASP (alice, bob, evil, etc.)",
					EnvVar: "RVASP_NAME",
				},
				cli.StringFlag{
					Name:   "a, addr",
					Usage:  "the address and port to bind the server on",
					Value:  ":4434",
					EnvVar: "RVASP_BIND_ADDR",
				},
				cli.StringFlag{
					Name:   "t, trisa-addr",
					Usage:  "the address and port to bind the TRISA server on",
					Value:  ":4435",
					EnvVar: "RVASP_TRISA_BIND_ADDR",
				},
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the dsn of the postgres database to connect to",
					EnvVar: "RVASP_DATABASE",
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
					Usage:  "the dsn of the postgres database to connect to",
					EnvVar: "RVASP_DATABASE",
				},
			},
		},
		{
			Name:     "resetdb",
			Usage:    "reset the database using the JSON fixtures",
			Category: "server",
			Action:   resetdb,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the dsn of the postgres database to connect to",
					EnvVar: "RVASP_DATABASE",
				},
				cli.StringFlag{
					Name:   "f, fixtures",
					Usage:  "the path to the fixtures directory",
					Value:  filepath.Join("pkg", "rvasp", "fixtures"),
					EnvVar: "RVASP_FIXTURES_PATH",
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
					Name:   "e, endpoint",
					Usage:  "the address and port to connect to the server on",
					Value:  "localhost:4434",
					EnvVar: "RVASP_ADDR",
				},
				cli.StringFlag{
					Name:   "a, account",
					Usage:  "the wallet or email address of the account",
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
					Name:   "e, endpoint",
					Usage:  "the address and port to connect to the server on",
					Value:  "localhost:4434",
					EnvVar: "RVASP_ADDR",
				},
				cli.StringFlag{
					Name:   "a, account",
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
				&cli.StringFlag{
					Name:  "B, beneficiary-vasp",
					Usage: "the common name or vasp directory searchable name of the beneficiary vasp",
				},
				&cli.BoolFlag{
					Name:  "E, external-demo",
					Usage: "whether the beneficiary is a demo node (e.g. alice or bob) or not",
				},
			},
		},
		{
			Name:     "stream",
			Usage:    "initiate a transfer stream for listening or initiating a transfer",
			Category: "client",
			Action:   stream,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "e, endpoint",
					Usage:  "the address and port to connect to the server on",
					Value:  "localhost:4434",
					EnvVar: "RVASP_ADDR",
				},
				cli.BoolFlag{
					Name:  "l, listen",
					Usage: "only listen for messages, do not send transfer",
				},
				cli.StringFlag{
					Name:   "a, account",
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
	var conf *rvasp.Settings
	if conf, err = rvasp.Config(); err != nil {
		return cli.NewExitError(err, 1)
	}

	if db := c.String("db"); db != "" {
		conf.DatabaseDSN = db
	}

	var db *gorm.DB
	if db, err = gorm.Open(postgres.Open(conf.DatabaseDSN), &gorm.Config{}); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = rvasp.MigrateDB(db); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

// Reset the database
func resetdb(c *cli.Context) (err error) {
	var conf *rvasp.Settings
	if conf, err = rvasp.Config(); err != nil {
		return cli.NewExitError(err, 1)
	}

	if db := c.String("db"); db != "" {
		conf.DatabaseDSN = db
	}

	if fixtures := c.String("fixtures"); fixtures != "" {
		conf.FixturesPath = fixtures
	}

	var db *gorm.DB
	if db, err = gorm.Open(postgres.Open(conf.DatabaseDSN), &gorm.Config{}); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = rvasp.ResetDB(db, conf.FixturesPath); err != nil {
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
		Account:         c.String("account"),
		Beneficiary:     c.String("beneficiary"),
		BeneficiaryVasp: c.String("beneficiary-vasp"),
		Amount:          float32(c.Float64("amount")),
		ExternalDemo:    c.Bool("external-demo"),
	}

	if req.Account == "" {
		return cli.NewExitError("specify account email", 1)
	}

	if req.Beneficiary == "" && req.BeneficiaryVasp == "" {
		return cli.NewExitError("specify a beneficiary or beneficiary vasp", 1)
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
		// Extract the status from the error
		var (
			ok   bool
			serr *status.Status
		)
		if serr, ok = status.FromError(err); !ok {
			return cli.NewExitError(err, 1)
		}
		fmt.Printf("[%d] %s\n", serr.Code(), serr.Message())
		return nil
	}
	return printJSON(rep)
}

// Client method: transfer funds
func stream(c *cli.Context) (err error) {
	var req *pb.TransferRequest
	if !c.Bool("listen") {
		req = &pb.TransferRequest{
			Account:          c.String("account"),
			Beneficiary:      c.String("beneficiary"),
			Amount:           float32(c.Float64("amount")),
			CheckBeneficiary: false,
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
	}

	client, err := makeDemoClient(c)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.LiveUpdates(ctx)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	// Send the client connection message
	if err = stream.Send(&pb.Command{
		Type:   pb.RPC_NORPC,
		Client: "testing",
	}); err != nil {
		return cli.NewExitError(err, 1)
	}

	if req != nil {
		if err = stream.Send(&pb.Command{
			Type:   pb.RPC_TRANSFER,
			Id:     1,
			Client: "testing",
			Request: &pb.Command_Transfer{
				Transfer: req,
			},
		}); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	for {
		select {
		case <-ctx.Done():
			if err = ctx.Err(); err != nil {
				return cli.NewExitError(err, 1)
			}
			return nil
		default:
		}

		msg, err := stream.Recv()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		printJSON(msg)
	}
}

// helper function to create the GRPC client with default options
func makeClient(c *cli.Context) (_ pb.TRISAIntegrationClient, err error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	var cc *grpc.ClientConn
	if cc, err = grpc.Dial(c.String("endpoint"), opts...); err != nil {
		return nil, err
	}
	return pb.NewTRISAIntegrationClient(cc), nil
}

func makeDemoClient(c *cli.Context) (_ pb.TRISADemoClient, err error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	var cc *grpc.ClientConn
	if cc, err = grpc.Dial(c.String("endpoint"), opts...); err != nil {
		return nil, err
	}
	return pb.NewTRISADemoClient(cc), nil
}

// helper function to print JSON response and exit
func printJSON(v protoreflect.ProtoMessage) error {
	jsonpb := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		AllowPartial:    true,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}

	data, err := jsonpb.Marshal(v)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	fmt.Println(string(data))
	return nil
}

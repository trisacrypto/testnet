package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/trisacrypto/testnet/pkg"
	openvasp "github.com/trisacrypto/testnet/pkg/openvasp/web-service-gin"
	"github.com/trisacrypto/testnet/pkg/rvasp/config"
	"github.com/trisacrypto/testnet/pkg/rvasp/db"
	"github.com/urfave/cli"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
					Name:   "a, addr",
					Usage:  "the address and port to bind the server on",
					Value:  ":4435",
					EnvVar: "RVASP_BIND_ADDR",
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
					EnvVar: "RVASP_DATABASE_DSN",
				},
				cli.BoolFlag{
					Name:  "L, no-load",
					Usage: "do not load initial fixtures into the database",
				},
				cli.StringFlag{
					Name:   "f, fixtures",
					Usage:  "the path to the fixtures directory to load into the database",
					Value:  filepath.Join("pkg", "rvasp", "fixtures"),
					EnvVar: "RVASP_FIXTURES_PATH",
				},
			},
		},
	}

	app.Run(os.Args)
}

// Serve the OpenVASP gin server
func serve(c *cli.Context) {
	router := gin.Default()
	router.POST("/register", openvasp.Register)
	router.POST("/transfer", openvasp.Transfer)
	router.Run(fmt.Sprintf("localhost%s", c.String("addr")))
}

// TODO: verify and test database initialization
// Run the database migration
func initdb(c *cli.Context) (err error) {

	var conf *config.Config
	if conf, err = config.New(); err != nil {
		return cli.NewExitError(err, 1)
	}

	if dsn := c.String("db"); dsn != "" {
		conf.Database.DSN = dsn
	}

	if conf.Database.DSN == "" {
		return cli.NewExitError("openvasp database dsn required", 1)
	}

	var gdb *gorm.DB
	if gdb, err = gorm.Open(postgres.Open(conf.Database.DSN), &gorm.Config{}); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = db.MigrateDB(gdb); err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}

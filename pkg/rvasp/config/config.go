package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

// Config uses envconfig to load required settings from the environment and
// validate them in preparation for running the rVASP.
//
// TODO: also store separate signing key instead of using the cert key.
type Config struct {
	Name           string          `envconfig:"RVASP_NAME"`
	BindAddr       string          `envconfig:"RVASP_BIND_ADDR" default:":4434"`
	TRISABindAddr  string          `envconfig:"RVASP_TRISA_BIND_ADDR" default:":4435"`
	FixturesPath   string          `envconfig:"RVASP_FIXTURES_PATH"`
	CertPath       string          `envconfig:"RVASP_CERT_PATH"`
	TrustChainPath string          `envconfig:"RVASP_TRUST_CHAIN_PATH"`
	AsyncInterval  time.Duration   `envconfig:"RVASP_ASYNC_INTERVAL" default:"1m"`
	AsyncNotBefore time.Duration   `envconfig:"RVASP_ASYNC_NOT_BEFORE" default:"5m"`
	AsyncNotAfter  time.Duration   `envconfig:"RVASP_ASYNC_NOT_AFTER" default:"1h"`
	ConsoleLog     bool            `envconfig:"RVASP_CONSOLE_LOG" default:"false"`
	LogLevel       LogLevelDecoder `envconfig:"RVASP_LOG_LEVEL" default:"info"`
	GDS            GDSConfig
	Database       DatabaseConfig
}

// GDSConfig is the configuration for connecting to GDS
type GDSConfig struct {
	URL string `split_words:"true" default:"api.trisatest.net:443"`
}

// DatabaseConfig is the configuration for connecting to the RVASP database
type DatabaseConfig struct {
	DSN        string `split_words:"true"`
	MaxRetries int    `split_words:"true" default:"0"`
}

// New creates a new Config object, loading environment variables and defaults.
func New() (_ *Config, err error) {
	var conf Config
	if err = envconfig.Process("rvasp", &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// LogLevelDecoder deserializes the log level from a config string.
type LogLevelDecoder zerolog.Level

// Decode implements envconfig.Decoder
func (ll *LogLevelDecoder) Decode(value string) error {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "panic":
		*ll = LogLevelDecoder(zerolog.PanicLevel)
	case "fatal":
		*ll = LogLevelDecoder(zerolog.FatalLevel)
	case "error":
		*ll = LogLevelDecoder(zerolog.ErrorLevel)
	case "warn":
		*ll = LogLevelDecoder(zerolog.WarnLevel)
	case "info":
		*ll = LogLevelDecoder(zerolog.InfoLevel)
	case "debug":
		*ll = LogLevelDecoder(zerolog.DebugLevel)
	case "trace":
		*ll = LogLevelDecoder(zerolog.TraceLevel)
	default:
		return fmt.Errorf("unknown log level %q", value)
	}
	return nil
}

package rvasp

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

// Settings uses envconfig to load required settings from the environment and
// validate them in preparation for running the rVASP.
//
// TODO: also store separate signing key instead of using the cert key.
type Settings struct {
	Name                string          `envconfig:"RVASP_NAME"`
	BindAddr            string          `envconfig:"RVASP_BIND_ADDR" default:":4434"`
	TRISABindAddr       string          `envconfig:"RVASP_TRISA_BIND_ADDR" default:":4435"`
	DatabaseDSN         string          `envconfig:"RVASP_DATABASE"`
	CertPath            string          `envconfig:"RVASP_CERT_PATH"`
	TrustChainPath      string          `envconfig:"RVASP_TRUST_CHAIN_PATH"`
	DirectoryServiceURL string          `envconfig:"RVASP_DIRECTORY_SERVICE_URL" default:"api.vaspdirectory.net:443"`
	ConsoleLog          bool            `envconfig:"RVASP_CONSOLE_LOG" default:"false"`
	LogLevel            LogLevelDecoder `envconfig:"RVASP_LOG_LEVEL" default:"info"`
}

// Config creates a new settings object, loading environment variables and defaults.
func Config() (_ *Settings, err error) {
	var conf Settings
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

package config

import (
	"log"

	"github.com/fishbrain/logging-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

var Config = getEnv()
var LoggingConfig logging.LoggingConfig

var VERSION string

// config represents the configuration of Bonito read from environment
type config struct {
	Environment   		string `envconfig:"BONITO_ENV" default:"development"`
	LogLevel      		string `envconfig:"LOG_LEVEL" default:"INFO"`
	LogGorp       		bool   `envconfig:"LOG_GORP" default:"false"`
	BugsnagAPIKey 		string `envconfig:"BUGSNAG_API_KEY"`
	GcpProdProjectId 	string `envconfig:"GCP_PROD_PROJECT_ID" default:""`
}

// GetEnv reads configuration from environment
func getEnv() config {
	var conf config
	err := envconfig.Process("bonito", &conf)
	if err != nil {
		log.Fatalf("failed to parse configuration (%s)", err)
	}

	LoggingConfig = logging.LoggingConfig{
		LogLevel:                   conf.LogLevel,
		Environment:                conf.Environment,
		AppVersion:                 VERSION,
		BugsnagAPIKey:              conf.BugsnagAPIKey,
		BugsnagNotifyReleaseStages: []string{"production", "staging"},
		BugsnagProjectPackages:     []string{"github.com/fishbrain/logging-go"},
		BugsnagProjectPaths:        []string{"/bonito"},
	}

	return conf
}

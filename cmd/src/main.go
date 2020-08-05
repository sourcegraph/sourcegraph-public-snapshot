package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const usageText = `src is a tool that provides access to Sourcegraph instances.
For more information, see https://github.com/sourcegraph/src-cli

Usage:

	src [options] command [command options]

Environment variables
	SRC_ACCESS_TOKEN  Sourcegraph access token
	SRC_ENDPOINT      endpoint to use, if unset will default to "https://sourcegraph.com"

The options are:

	-v                               print verbose output

The commands are:

	search          search for results on Sourcegraph
	api             interacts with the Sourcegraph GraphQL API
	repos,repo      manages repositories
	users,user      manages users
	orgs,org        manages organizations
	config          manages global, org, and user settings
	extsvc          manages external services
	extensions,ext  manages extensions (experimental)
	actions         runs actions to generate patch sets (experimental)
	campaigns       manages campaigns (experimental)
	lsif            manages LSIF data
	serve-git       serves your local git repositories over HTTP for Sourcegraph to pull
	version         display and compare the src-cli version against the recommended version for your instance

Use "src [command] -h" for more information about a command.

`

var (
	verbose = flag.Bool("v", false, "print verbose output")

	// The following arguments are deprecated which is why they are no longer documented
	configPath = flag.String("config", "", "")
	endpoint   = flag.String("endpoint", "", "")
)

// commands contains all registered subcommands.
var commands commander

func main() {
	// Configure logging.
	log.SetFlags(0)
	log.SetPrefix("")

	commands.run(flag.CommandLine, "src", usageText, os.Args[1:])
}

var cfg *config

// config represents the config format.
type config struct {
	Endpoint          string            `json:"endpoint"`
	AccessToken       string            `json:"accessToken"`
	AdditionalHeaders map[string]string `json:"additionalHeaders"`
}

// readConfig reads the config file from the given path.
func readConfig() (*config, error) {
	cfgPath := *configPath
	userSpecified := *configPath != ""

	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	if !userSpecified {
		cfgPath = filepath.Join(u.HomeDir, "src-config.json")
	} else if strings.HasPrefix(cfgPath, "~/") {
		cfgPath = filepath.Join(u.HomeDir, cfgPath[2:])
	}
	data, err := ioutil.ReadFile(os.ExpandEnv(cfgPath))
	if err != nil && (!os.IsNotExist(err) || userSpecified) {
		return nil, err
	}
	var cfg config
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	envToken := os.Getenv("SRC_ACCESS_TOKEN")
	envEndpoint := os.Getenv("SRC_ENDPOINT")

	if userSpecified {
		// If a config file is present, either zero or both environment variables must be present.
		// We don't want to partially apply environment variables.
		if envToken == "" && envEndpoint != "" {
			return nil, errConfigMerge
		}
		if envToken != "" && envEndpoint == "" {
			return nil, errConfigMerge
		}
	}

	// Apply config overrides.
	if envToken != "" {
		cfg.AccessToken = envToken
	}
	if envEndpoint != "" {
		cfg.Endpoint = envEndpoint
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://sourcegraph.com"
	}

	cfg.AdditionalHeaders = parseAdditionalHeaders()

	// Lastly, apply endpoint flag if set
	if endpoint != nil && *endpoint != "" {
		cfg.Endpoint = *endpoint
	}

	cfg.Endpoint = strings.TrimSuffix(cfg.Endpoint, "/")

	return &cfg, nil
}

var errConfigMerge = errors.New("when using a configuration file, zero or all environment variables must be set")

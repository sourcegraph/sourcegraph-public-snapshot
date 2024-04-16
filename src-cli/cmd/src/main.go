package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
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

	api             interacts with the Sourcegraph GraphQL API
	batch           manages batch changes
	code-intel      manages code intelligence data
	config          manages global, org, and user settings
	extensions,ext  manages extensions (experimental)
	extsvc          manages external services
	login           authenticate to a Sourcegraph instance with your user credentials
	lsif            manages LSIF data (deprecated: use 'code-intel')
	orgs,org        manages organizations
	teams,team      manages teams
	repos,repo      manages repositories
	search          search for results on Sourcegraph
	serve-git       serves your local git repositories over HTTP for Sourcegraph to pull
	users,user      manages users
	codeowners      manages code ownership information
	version         display and compare the src-cli version against the recommended version for your instance

Use "src [command] -h" for more information about a command.

`

var (
	verbose = flag.Bool("v", false, "print verbose output")

	// The following arguments are deprecated which is why they are no longer documented
	configPath = flag.String("config", "", "")
	endpoint   = flag.String("endpoint", "", "")

	errConfigMerge                 = errors.New("when using a configuration file, zero or all environment variables must be set")
	errConfigAuthorizationConflict = errors.New("when passing an 'Authorization' additional headers, SRC_ACCESS_TOKEN must never be set")
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

	ConfigFilePath string
}

// apiClient returns an api.Client built from the configuration.
func (c *config) apiClient(flags *api.Flags, out io.Writer) api.Client {
	return api.NewClient(api.ClientOpts{
		Endpoint:          c.Endpoint,
		AccessToken:       c.AccessToken,
		AdditionalHeaders: c.AdditionalHeaders,
		Flags:             flags,
		Out:               out,
	})
}

var testHomeDir string // used by tests to mock the user's $HOME

// readConfig reads the config file from the given path.
func readConfig() (*config, error) {
	cfgPath := *configPath
	userSpecified := *configPath != ""

	var homeDir string
	if testHomeDir != "" {
		homeDir = testHomeDir
	} else {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		homeDir = u.HomeDir
	}

	if !userSpecified {
		cfgPath = filepath.Join(homeDir, "src-config.json")
	} else if strings.HasPrefix(cfgPath, "~/") {
		cfgPath = filepath.Join(homeDir, cfgPath[2:])
	}
	data, err := os.ReadFile(os.ExpandEnv(cfgPath))
	if err != nil && (!os.IsNotExist(err) || userSpecified) {
		return nil, err
	}
	var cfg config
	if err == nil {
		cfg.ConfigFilePath = cfgPath
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
	// Ensure that we're not clashing additonal headers
	_, hasAuthorizationAdditonalHeader := cfg.AdditionalHeaders["authorization"]
	if cfg.AccessToken != "" && hasAuthorizationAdditonalHeader {
		return nil, errConfigAuthorizationConflict
	}

	// Lastly, apply endpoint flag if set
	if endpoint != nil && *endpoint != "" {
		cfg.Endpoint = *endpoint
	}

	cfg.Endpoint = cleanEndpoint(cfg.Endpoint)

	return &cfg, nil
}

func cleanEndpoint(urlStr string) string {
	return strings.TrimSuffix(urlStr, "/")
}

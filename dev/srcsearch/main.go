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
)

const usageText = `srcsearch runs a search against a Sourcegraph instance.

Usage:

	srcsearch [options] query

The options are:

	-config=$HOME/src-config.json    specifies a file containing {"accessToken": "<secret>", "endpoint": "https://sourcegraph.com"}
	-endpoint=                       specifies the endpoint to use e.g. "https://sourcegraph.com" (overrides -config, if any)

`

var (
	configPath = flag.String("config", "", "")
	endpoint   = flag.String("endpoint", "", "")
)

// commands contains all registered subcommands.
var commands commander

func main() {
	// Configure logging.
	log.SetFlags(0)
	log.SetPrefix("")

	commands.run(flag.CommandLine, "srcsearch", usageText, os.Args[1:])
}

var cfg *config

// config represents the config format.
type config struct {
	Endpoint    string `json:"endpoint"`
	AccessToken string `json:"accessToken"`
}

// readConfig reads the config file from the given path.
func readConfig() (*config, error) {
	cfgPath := *configPath
	userSpecified := *configPath != ""
	if !userSpecified {
		user, err := user.Current()
		if err != nil {
			return nil, err
		}
		cfgPath = filepath.Join(user.HomeDir, "src-config.json")
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

	// Apply config overrides.
	if envToken := os.Getenv("SRC_ACCESS_TOKEN"); envToken != "" {
		cfg.AccessToken = envToken
	}
	if *endpoint != "" {
		cfg.Endpoint = strings.TrimSuffix(*endpoint, "/")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://sourcegraph.com"
	}
	return &cfg, nil
}

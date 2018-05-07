package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const usageText = `src is a tool that provides access to Sourcegraph instances.
For more information, see https://github.com/sourcegraph/src-cli

Usage:

	src [options] command [command options]

The options are:

	-config=$HOME/src-config.json    specifies a file containing {"accessToken": "<secret>", "endpoint": "https://sourcegraph.com"}
	-endpoint=                       specify the endpoint to use e.g. "https://sourcegraph.com" (overrides -config, if any)

The commands are:

	api    interact with the Sourcegraph GraphQL API

Use "src [command] -h" for more information about a command.

`

var (
	configPath = flag.String("config", "", "")
	endpoint   = flag.String("endpoint", "", "")
)

// command is a subcommand handler and its flag set.
type command struct {
	// flagSet is the flag set for the command.
	flagSet *flag.FlagSet

	// aliases for the command.
	aliases []string

	// handler is the function that is invoked to handle this command.
	handler func(args []string) error

	// flagSet.Usage function to invoke on e.g. -h flag. If nil, a default one
	// one is used.
	usageFunc func()
}

// matches tells if the given name matches this command or one of its aliases.
func (c *command) matches(name string) bool {
	if name == c.flagSet.Name() {
		return true
	}
	for _, alias := range c.aliases {
		if name == alias {
			return true
		}
	}
	return false
}

// commands contains all registered subcommands.
var commands []*command

func main() {
	// Configure logging.
	log.SetFlags(0)
	log.SetPrefix("")

	// Configure usage.
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageText)
	}

	// Parse flags.
	flag.Parse()

	// Print usage if the command is "help".
	if flag.Arg(0) == "help" || flag.NArg() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// Configure default usage funcs for commands.
	for _, cmd := range commands {
		if cmd.usageFunc != nil {
			cmd.flagSet.Usage = cmd.usageFunc
			continue
		}
		cmd.flagSet.Usage = func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src %s':\n", cmd.flagSet.Name())
			cmd.flagSet.PrintDefaults()
		}
	}

	// Find the subcommand to execute.
	name := flag.Arg(0)
	for _, cmd := range commands {
		if !cmd.matches(name) {
			continue
		}

		// Read global configuration now.
		var err error
		cfg, err = readConfig()
		if err != nil {
			log.Fatal("reading config: ", err)
		}

		// Parse subcommand flags.
		args := flag.Args()[1:]
		if err := cmd.flagSet.Parse(args); err != nil {
			panic(fmt.Sprintf("all registered commands should use flag.ExitOnError: error: %s", err))
		}

		// Execute the subcommand.
		if err := cmd.handler(flag.Args()[1:]); err != nil {
			if _, ok := err.(*usageError); ok {
				log.Println(err)
				cmd.flagSet.Usage()
				os.Exit(2)
			}
			log.Fatal(err)
		}
		os.Exit(0)
	}
	log.Printf("src: unknown subcommand %q", name)
	log.Fatal("Run 'src help' for usage.")
}

// usageError is an error type that subcommands can return in order to signal
// that a usage error has occurred.
type usageError struct {
	error
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
		cfg.Endpoint = *endpoint
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://sourcegraph.com"
	}

	if cfg.AccessToken == "" {
		return nil, fmt.Errorf(`error: you must specify an access token to use for %s

You can do so via the environment:

	SRC_ACCESS_TOKEN="secret" src ...

or via the configuration file (%s):

	{"accessToken": "secret"}

`, cfg.Endpoint, cfgPath)
	}
	return &cfg, nil
}

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/validate/install"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	usage := `'src validate install' is a tool that validates a Sourcegraph installation.

Examples:

	Run default checks:

		$ src validate install

	Provide a YAML configuration file:

		$ src validate install config.yml

	Provide a JSON configuration file:

		$ src validate install config.json

Environmental variables

	SRC_GITHUB_TOKEN	GitHub access token for validation features

`

	flagSet := flag.NewFlagSet("install", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src validate %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		var validationSpec *install.ValidationSpec

		if len(flagSet.Args()) == 1 {
			fileName := flagSet.Arg(0)
			f, err := os.ReadFile(fileName)
			if err != nil {
				return errors.Wrap(err, "failed to read installation validation config from file")
			}

			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				validationSpec, err = install.LoadYamlConfig(f)
				if err != nil {
					return err
				}
			}

			if strings.HasSuffix(fileName, ".json") {
				validationSpec, err = install.LoadJsonConfig(f)
				if err != nil {
					return err
				}
			}
		}

		if !isatty.IsTerminal(os.Stdin.Fd()) {
			// stdin is a pipe not a terminal
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return errors.Wrap(err, "failed to read installation validation config from pipe")
			}
			validationSpec, err = install.LoadYamlConfig(input)
			if err != nil {
				return err
			}
		}

		if validationSpec == nil {
			validationSpec = install.DefaultConfig()
		}

		validationSpec.ExternalService.Config.GitHub.Token = os.Getenv("SRC_GITHUB_TOKEN")

		return install.Validate(context.Background(), client, validationSpec)
	}

	validateCommands = append(validateCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

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
	"github.com/sourcegraph/src-cli/internal/validate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	validateWarningEmoji = output.EmojiWarning
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

		var validationSpec *validate.ValidationSpec

		if len(flagSet.Args()) == 1 {
			fileName := flagSet.Arg(0)
			f, err := os.ReadFile(fileName)
			if err != nil {
				return errors.Wrap(err, "failed to read installation validation config from file")
			}

			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				validationSpec, err = validate.LoadYamlConfig(f)
				if err != nil {
					return err
				}
			}

			if strings.HasSuffix(fileName, ".json") {
				validationSpec, err = validate.LoadJsonConfig(f)
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
			validationSpec, err = validate.LoadYamlConfig(input)
			if err != nil {
				return err
			}
		}

		if validationSpec == nil {
			validationSpec = validate.DefaultConfig()
		}

		envGithubToken := os.Getenv("SRC_GITHUB_TOKEN")

		// will work for now with only GitHub supported but will need to be revisited when other code hosts are supported
		if envGithubToken == "" {
			return errors.Newf("%s failed to read `SRC_GITHUB_TOKEN` environment variable", validateWarningEmoji)
		}

		validationSpec.ExternalService.Config.GitHub.Token = envGithubToken

		return validate.Installation(context.Background(), client, validationSpec)
	}

	validateCommands = append(validateCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

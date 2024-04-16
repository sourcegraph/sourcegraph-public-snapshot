package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
  Examples:

  create an external service configuration on the Sourcegraph instance:

  $ cat new-config.json | src extsvc create
  $ src extsvc create -name 'My GitHub connection' new-config.json
  `

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extsvc %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag = flagSet.String("name", "", "exact name of the external service to create")
		kindFlag = flagSet.String("kind", "", "kind of the external service to create")
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) (err error) {
		ctx := context.Background()
		if err := flagSet.Parse(args); err != nil {
			return err
		}
		if *nameFlag == "" {
			return errors.New("-name must be provided")
		}

		var createJSON []byte
		if len(flagSet.Args()) == 1 {
			createJSON, err = os.ReadFile(flagSet.Arg(0))
			if err != nil {
				return err
			}
		}
		if !isatty.IsTerminal(os.Stdin.Fd()) {
			// stdin is a pipe not a terminal
			createJSON, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		}

		createExternalServiceInput := map[string]interface{}{
			"kind":        strings.ToUpper(*kindFlag),
			"displayName": *nameFlag,
			"config":      string(createJSON),
		}
		queryVars := map[string]interface{}{
			"input": createExternalServiceInput,
		}
		var result struct{} // TODO: future: allow formatting resulting external service

		client := cfg.apiClient(apiFlags, flagSet.Output())
		if ok, err := client.NewRequest(externalServicesCreateMutation, queryVars).Do(ctx, &result); err != nil {
			return err
		} else if ok {
			fmt.Println("External service created:", *nameFlag)
		}
		return nil
	}

	// Register the command.
	extsvcCommands = append(extsvcCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const externalServicesCreateMutation = `
  mutation AddExternalService($input: AddExternalServiceInput!) {
    addExternalService(input: $input) {
      id
      warning
    }
  }`

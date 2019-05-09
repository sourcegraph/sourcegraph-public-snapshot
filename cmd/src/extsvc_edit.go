package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	isatty "github.com/mattn/go-isatty"
)

func init() {
	usage := `
Examples:

  Edit an external service configuration on the Sourcegraph instance:

    	$ cat new-config.json | src extsvc edit -id 'RXh0ZXJuYWxTZXJ2aWNlOjQ='
    	$ src extsvc edit -name 'My GitHub connection' new-config.json

  Edit an external service name on the Sourcegraph instance:

    	$ src extsvc edit -name 'My GitHub connection' -rename 'New name'

`

	flagSet := flag.NewFlagSet("edit", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extsvc %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag   = flagSet.String("name", "", "exact name of the external service to edit")
		idFlag     = flagSet.String("id", "", "ID of the external service to edit")
		renameFlag = flagSet.String("rename", "", "when specified, renames the external service")
		apiFlags   = newAPIFlags(flagSet)
	)

	handler := func(args []string) (err error) {
		flagSet.Parse(args)

		// Determine ID of external service we will edit.
		if *nameFlag == "" && *idFlag == "" {
			return &usageError{errors.New("one of -name or -id flag must be specified")}
		}
		id := *idFlag
		if id == "" {
			id, err = lookupExternalServiceByName(*nameFlag)
			if err != nil {
				return err
			}
		}

		// Determine if we are updating the JSON configuration or not.
		var updateJSON []byte
		if len(flagSet.Args()) == 1 {
			updateJSON, err = ioutil.ReadFile(flagSet.Arg(0))
			if err != nil {
				return err
			}
		}
		if !isatty.IsTerminal(os.Stdin.Fd()) {
			// stdin is a pipe not a terminal
			updateJSON, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		}

		updateExternalServiceInput := map[string]interface{}{
			"id": id,
		}
		if *renameFlag != "" {
			updateExternalServiceInput["displayName"] = *renameFlag
		}
		if len(updateJSON) > 0 {
			updateExternalServiceInput["config"] = string(updateJSON)
		}
		if len(updateExternalServiceInput) == 1 {
			return nil // nothing to update
		}

		queryVars := map[string]interface{}{
			"input": updateExternalServiceInput,
		}
		var result struct{} // TODO: future: allow formatting resulting external service
		return (&apiRequest{
			query:  externalServicesUpdateMutation,
			vars:   queryVars,
			result: &result,
			done: func() error {
				fmt.Println("External service updated:", id)
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	extsvcCommands = append(extsvcCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const externalServicesUpdateMutation = `
mutation ($input: UpdateExternalServiceInput!) {
	updateExternalService(input: $input) {
		id
	}
}
`

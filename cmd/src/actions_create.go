package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const actionDefinitionTemplate = `{
  "$schema": "https://github.com/sourcegraph/src-cli/tree/master/schema/actions.schema.json",
  "scopeQuery": "",
  "steps": [
  ]
}
`

func init() {
	usage := `
Create an empty action definition in action.json (if not -o flag is given). This command is meant to help with creating action definitions to be used with 'src actions exec'.

Examples:

  Create a new action definition in action.json:

		$ src actions create

  Create a new action definition in ~/Documents/my-action.json:

		$ src actions create -o ~/Documents/my-action.json
`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src actions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}

	var (
		fileFlag = flagSet.String("o", "action.json", "The destination file name. Default value is 'action.json'")
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		if _, err := os.Stat(*fileFlag); !os.IsNotExist(err) {
			return fmt.Errorf("file %q already exists", *fileFlag)
		}

		return ioutil.WriteFile(*fileFlag, []byte(actionDefinitionTemplate), 0644)
	}

	// Register the command.
	actionsCommands = append(actionsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

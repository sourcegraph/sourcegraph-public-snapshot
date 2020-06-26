package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const actionDefinitionTemplate = `scopeQuery: ""

steps:
  - type: command
    args:
    - echo
    - Hello world
`

func init() {
	usage := `
Create an empty action definition in action.yml (if no -o flag is given). This command is meant to help with creating action definitions to be used with 'src actions exec'.

Examples:

  Create a new action definition in action.yml:

		$ src actions create

  Create a new action definition in ~/Documents/my-action.yml:

		$ src actions create -o ~/Documents/my-action.yml
`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src actions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}

	var (
		fileFlag = flagSet.String("o", "action.yml", "The destination file name. Default value is 'action.yml'")
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

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

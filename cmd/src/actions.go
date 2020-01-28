package main

import (
	"flag"
	"fmt"
)

var actionsCommands commander

func init() {
	usage := `'src actions' is a tool that manages actions on a Sourcegraph instance.

EXPERIMENTAL: Actions are experimental functionality on Sourcegraph and in the 'src' tool.

Usage:

	src actions command [command options]

The commands are:

	exec              executes an action to produce patches
	scope-query       list the repositories matched by "scopeQuery" in action

Use "src actions [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("actions", flag.ExitOnError)
	handler := func(args []string) error {
		actionsCommands.run(flagSet, "src actions", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"action"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

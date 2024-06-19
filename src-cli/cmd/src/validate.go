package main

import (
	"flag"
	"fmt"
)

var validateCommands commander

func init() {
	usage := `'src validate' is a tool that validates a Sourcegraph instance.

EXPERIMENTAL: 'validate' is an experimental command in the 'src' tool.

Please visit https://docs.sourcegraph.com/admin/validation for documentation of the validate command.

Usage:

	src validate command [command options]

The commands are:

	install        validates a Sourcegraph installation
	kube           validates a Sourcegraph deployment on a Kubernetes cluster

Use "src validate [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("validate", flag.ExitOnError)
	handler := func(args []string) error {
		validateCommands.run(flagSet, "src validate", usage, args)
		return nil
	}

	// Register the command
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"validate"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

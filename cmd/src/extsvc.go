package main

import (
	"flag"
	"fmt"
)

var extsvcCommands commander

func init() {
	usage := `'src extsvc' is a tool that manages external services on a Sourcegraph instance.

Usage:

	src extsvc command [command options]

The commands are:

	list      lists the external services on the Sourcegraph instance

Use "src extsvc [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("extsvc", flag.ExitOnError)
	handler := func(args []string) error {
		extsvcCommands.run(flagSet, "src extsvc", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"extsvc", "external-service"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

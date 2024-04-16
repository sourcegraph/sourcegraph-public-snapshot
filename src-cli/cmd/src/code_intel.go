package main

import (
	"flag"
	"fmt"
)

var codeintelCommands commander

func init() {
	usage := `'src code-intel' manages code intelligence data on a Sourcegraph instance.

Usage:

    src code-intel command [command options]

The commands are:

    upload     uploads a SCIP or LSIF index

Use "src code-intel [command] -h" for more information about a command.
`
	flagSet := flag.NewFlagSet("code-intel", flag.ExitOnError)
	handler := func(args []string) error {
		lsifCommands.run(flagSet, "src code-intel", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"code-intel"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

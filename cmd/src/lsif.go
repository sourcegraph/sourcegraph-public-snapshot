package main

import (
	"flag"
	"fmt"
)

var lsifCommands commander

func init() {
	usage := `'src lsif' is a tool that manages LSIF data on a Sourcegraph instance.

Usage:

	src lsif command [command options]

The commands are:

	upload     uploads an LSIF dump file

Use "src lsif [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("lsif", flag.ExitOnError)
	handler := func(args []string) error {
		lsifCommands.run(flagSet, "src lsif", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"lsif"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

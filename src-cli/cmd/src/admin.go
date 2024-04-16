package main

import (
	"flag"
	"fmt"
)

var adminCommands commander

func init() {
	usage := `'src admin' is a tool that manages an initial admin user on a new Sourcegraph instance.

Usage:
	
	src admin create [command options]

The commands are:

	create		create an initial admin user

Use "src admin [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("admin", flag.ExitOnError)
	handler := func(args []string) error {
		adminCommands.run(flagSet, "srv admin", usage, args)
		return nil
	}

	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"admin"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

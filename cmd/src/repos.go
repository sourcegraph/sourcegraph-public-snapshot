package main

import (
	"flag"
	"fmt"
)

var reposCommands commander

func init() {
	usage := `'src repos' is a tool that manages repositories on a Sourcegraph instance.

Usage:

	src repos command [command options]

The commands are:

	list       lists repositories
	enable     enables repositories
	disable    disables repositories
	delete 	   deletes repositories

Use "src repos [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("repos", flag.ExitOnError)
	handler := func(args []string) error {
		reposCommands.run(flagSet, "src repos", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"repo"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

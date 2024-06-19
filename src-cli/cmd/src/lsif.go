package main

import (
	"flag"
	"fmt"
)

var lsifCommands commander

func init() {
	usage := `[DEPRECATED] 'src lsif' is a tool that manages LSIF data on a Sourcegraph instance.

Use 'src code-intel' instead.
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

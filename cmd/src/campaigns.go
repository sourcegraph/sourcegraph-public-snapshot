package main

import (
	"flag"
	"fmt"
)

var campaignsCommands commander

func init() {
	usage := `'src campaigns' is a tool that manages campaigns on a Sourcegraph instance.

Usage:

	src campaigns command [command options]

The commands are:

	apply                 applies a campaign spec to create or update a campaign
	preview               creates a campaign spec to be previewed or applied
	repos,repositories    queries the exact repositories that a campaign spec
	                      will apply to
	validate              validates a campaign spec

Use "src campaigns [command] -h" for more information about a command.

`

	flagSet := flag.NewFlagSet("campaigns", flag.ExitOnError)
	handler := func(args []string) error {
		campaignsCommands.run(flagSet, "src campaigns", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet:   flagSet,
		aliases:   []string{"campaign"},
		handler:   handler,
		usageFunc: func() { fmt.Println(usage) },
	})
}

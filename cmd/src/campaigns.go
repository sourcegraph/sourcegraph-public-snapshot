package main

import (
	"flag"
	"fmt"
)

var campaignsCommands commander

func init() {
	usage := `'src campaigns' is a tool that manages campaigns on a Sourcegraph instance.

EXPERIMENTAL: Campaigns are experimental functionality on Sourcegraph and in the 'src' tool.

Usage:

	src campaigns command [command options]

The commands are:

	list              lists campaigns
	add-changesets    add changesets of a given repository to a campaign

Use "src campaigns [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("campaigns", flag.ExitOnError)
	handler := func(args []string) error {
		campaignsCommands.run(flagSet, "src campaigns", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

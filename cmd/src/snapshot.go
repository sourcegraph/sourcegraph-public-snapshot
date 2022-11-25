package main

import (
	"flag"
	"fmt"
)

var snapshotCommands commander

func init() {
	usage := `'src snapshot' manages snapshots of Sourcegraph instance data. All subcommands are currently EXPERIMENTAL.

USAGE
	src [-v] snapshot <command>

COMMANDS

	summary   export summary data about an instance for acceptance testing of a restored Sourcegraph instance
	test      use exported summary data and instance health indicators to validate a restored and upgraded instance
`
	flagSet := flag.NewFlagSet("snapshot", flag.ExitOnError)

	commands = append(commands, &command{
		flagSet: flagSet,
		handler: func(args []string) error {
			snapshotCommands.run(flagSet, "src snapshot", usage, args)
			return nil
		},
		usageFunc: func() { fmt.Fprint(flag.CommandLine.Output(), usage) },
	})
}

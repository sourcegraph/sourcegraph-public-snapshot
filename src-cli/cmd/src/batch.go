package main

import (
	"flag"
	"fmt"
)

var batchCommands commander

func init() {
	usage := `'src batch' manages batch changes on a Sourcegraph instance.

Usage:

	src batch command [command options]

The commands are:

	apply                 applies a batch spec to create or update a batch
	                      change
	new                   creates a new batch spec YAML file
	preview               creates a batch spec to be previewed or applied
	remote                creates server side batch changes
	repos,repositories    queries the exact repositories that a batch spec will
	                      apply to
	validate              validates a batch spec

Use "src batch [command] -h" for more information about a command.

`

	flagSet := flag.NewFlagSet("batch", flag.ExitOnError)
	handler := func(args []string) error {
		batchCommands.run(flagSet, "src batch", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{
			"batchchange",
			"batch-change",
			"batchchanges",
			"batch-changes",
			"batches",
		},
		handler:   handler,
		usageFunc: func() { fmt.Println(usage) },
	})
}

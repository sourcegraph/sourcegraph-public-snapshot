package main

import (
	"flag"
	"fmt"
)

var debugCommands commander

func init() {
	usage := `'src debug' gathers and bundles debug data from a Sourcegraph deployment for troubleshooting.

Usage:

	src debug command [command options]

The commands are:

	kube                 dumps context from k8s deployments
	compose              dumps context from docker-compose deployments
	server               dumps context from single-container deployments
	

Use "src debug command -h" for more information about a subcommands.
src debug has access to flags on src -- Ex: src -v kube -o foo.zip

`

	flagSet := flag.NewFlagSet("debug", flag.ExitOnError)
	handler := func(args []string) error {
		debugCommands.run(flagSet, "src debug", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet:   flagSet,
		aliases:   []string{},
		handler:   handler,
		usageFunc: func() { fmt.Println(usage) },
	})
}

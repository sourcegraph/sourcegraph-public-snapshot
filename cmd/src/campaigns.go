package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

var campaignsCommands commander

func init() {
	usage := `'src campaigns' is a tool that manages campaigns on a Sourcegraph instance.

Usage:

	src campaigns command [command options]

The commands are:

	apply                 applies a campaign spec to create or update a campaign
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

func campaignsOpenFileFlag(flag *string) (io.ReadCloser, error) {
	if flag == nil || *flag == "" || *flag == "-" {
		return os.Stdin, nil
	}

	file, err := os.Open(*flag)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %q", *flag)
	}
	return file, nil
}

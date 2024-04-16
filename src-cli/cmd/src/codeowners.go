package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var codeownersCommands commander

func init() {
	usage := `'src codeowners' is a tool that manages ingested code ownership data in a Sourcegraph instance.

Usage:

	src codeowners command [command options]

The commands are:

	get	returns the codeowners file for a repository, if exists
	create	create a codeowners file
	update	update a codeowners file
	delete	delete a codeowners file

Use "src codeowners [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("codeowners", flag.ExitOnError)
	handler := func(args []string) error {
		codeownersCommands.run(flagSet, "src codeowners", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"codeowner"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

const codeownersFragment = `
fragment CodeownersFileFields on CodeownersIngestedFile {
    contents
    repository {
		name
	}
}
`

type CodeownersIngestedFile struct {
	Contents   string `json:"contents"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func readFile(f string) (io.Reader, error) {
	if f == "-" {
		return os.Stdin, nil
	}
	return os.Open(f)
}

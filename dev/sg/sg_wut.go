package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var wutCommand = &cli.Command{
	Name:  "wut",
	Usage: "View, list, search and add abreviations used in Sourcegraph",
	UsageText: `
# List
`,
	Category: CategoryCompany,
	// This is where we do the actual searching for an abbreviation.
	// Assuming the command entered is `sg wut RFC`, this goes ahead and searches
	// the glossary for RFC.
	Action: func(ctx *cli.Context) error {
		if len(os.Args) == 2 {
			return errors.New("No search term provided")
		}

		searchTerm := os.Args[2]
		fmt.Printf("Searching for %s\n", searchTerm)
		return nil
	},
	// These are the subcommands available under the `sg wut` command.
	// - sg wut add
	// - sg wut refresh
	// - sg wut list
	// e.t.c
	Subcommands: []*cli.Command{
		{
			Name:      "add",
			ArgsUsage: " ",
			Usage:     "Add an abbreviation to the glossary",
			Action: func(ctx *cli.Context) error {
				fmt.Println("executing add subcommand")
				return nil
			},
		},
		{
			Name:      "refresh",
			ArgsUsage: " ",
			Usage:     "Refresh local glossary cache",
			Action: func(ctx *cli.Context) error {
				fmt.Println("executing refresh subcommand")
				return nil
			},
		},
	},
}

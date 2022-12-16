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
	Action: func(ctx *cli.Context) error {
		if len(os.Args) == 2 {
			return errors.New("No search term provided")
		}

		searchTerm := os.Args[2]
		fmt.Printf("Searching for %s\n", searchTerm)
		return nil
	},
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

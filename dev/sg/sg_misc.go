package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

var miscCommand = &cli.Command{
	Name:     "misc",
	Usage:    "Misc utilities",
	Category: CategoryUtil,
	Subcommands: []*cli.Command{{
		Name: "str2regex",
		Usage: `Turn a string into a regexp that Thorsten can paste in Sourcegraph search
strings.Split(regexp.QuoteMeta(str), " ") 👉 foo\s+:=\s+internal\.Get\(something\)`,
		Category:  CategoryUtil,
		UsageText: `sg misc str2regex 'Horsegraph is (very) cool'`,
		Action: func(cmd *cli.Context) error {
			str := strings.Join(cmd.Args().Slice(), " ")
			words := strings.Split(regexp.QuoteMeta(str), " ")
			fmt.Printf(strings.Join(words, `\s+`))
			return nil
		},
	}},
}

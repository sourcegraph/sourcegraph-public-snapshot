package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var helpCommand = &cli.Command{
	Name:      "help",
	ArgsUsage: " ", // no args accepted for now
	Usage:     "Get help about sg",
	Category:  CategoryUtil,
	Action: func(cmd *cli.Context) error {
		if cmd.NArg() != 0 {
			return errors.Newf("unexpected argument %s", cmd.Args().First())
		}
		cli.ShowAppHelpAndExit(cmd, 0)
		return nil
	},
}

package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/rfc"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var rfcCommand = &cli.Command{
	Name:  "rfc",
	Usage: `List, search, and open Sourcegraph RFCs`,
	UsageText: `
# List all Public RFCs
sg rfc list

# List all Private RFCs
sg rfc --private list

# Search for a Public RFC
sg rfc search "search terms"

# Search for a Private RFC
sg rfc --private search "search terms"

# Open a specific Public RFC
sg rfc open 420

# Open a specific private RFC
sg rfc --private open 420
`,
	Category: CategoryCompany,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:     "private",
			Usage:    "perform the rfc action on the private RFC drive",
			Required: false,
			Value:    false,
		},
	},
	Action: rfcExec,
}

func rfcExec(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		args = append(args, "list")
	}

	driveSpec := rfc.PublicDrive
	if ctx.Bool("private") {
		driveSpec = rfc.PrivateDrive
	}

	switch args[0] {
	case "list":
		return rfc.List(ctx.Context, driveSpec, std.Out)

	case "search":
		if len(args) != 2 {
			return errors.New("no search query given")
		}

		return rfc.Search(ctx.Context, args[1], driveSpec, std.Out)

	case "open":
		if len(args) != 2 {
			return errors.New("no number given")
		}

		return rfc.Open(ctx.Context, args[1], driveSpec, std.Out)
	}

	return nil
}

package main

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/rfc"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var rfcCommand = &cli.Command{
	Name:        "rfc",
	Usage:       "Run the given RFC command to manage RFCs",
	Description: `List, search and open Sourcegraph RFCs`,
	Category:    CategoryCompany,
	Action:      execAdapter(rfcExec),
}

func rfcExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		args = append(args, "list")
	}

	switch args[0] {
	case "list":
		return rfc.List(ctx, std.Out.Output)

	case "search":
		if len(args) != 2 {
			return errors.New("no search query given")
		}

		return rfc.Search(ctx, args[1], std.Out.Output)

	case "open":
		if len(args) != 2 {
			return errors.New("no number given")
		}

		return rfc.Open(ctx, args[1], std.Out.Output)
	}

	return nil
}

package main

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/rfc"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var rfcCommand = &cli.Command{
	Name:  "rfc",
	Usage: `List, search, and open Sourcegraph RFCs`,
	UsageText: `
# List all RFCs
sg rfc list

# Search for an RFC
sg rfc search "search terms"

# Open a specific RFC
sg rfc open 420
`,
	Category: CategoryCompany,
	Action:   execAdapter(rfcExec),
}

func rfcExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		args = append(args, "list")
	}

	switch args[0] {
	case "list":
		return rfc.List(ctx, std.Out)

	case "search":
		if len(args) != 2 {
			return errors.New("no search query given")
		}

		return rfc.Search(ctx, args[1], std.Out)

	case "open":
		if len(args) != 2 {
			return errors.New("no number given")
		}

		return rfc.Open(ctx, args[1], std.Out)
	}

	return nil
}

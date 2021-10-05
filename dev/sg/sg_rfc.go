package main

import (
	"context"
	"flag"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/rfc"
)

var (
	rfcFlagSet = flag.NewFlagSet("sg rfc", flag.ExitOnError)
	rfcCommand = &ffcli.Command{
		Name:       "rfc",
		ShortUsage: "sg rfc [list|search|open]",
		ShortHelp:  "Run the given RFC command to manage RFCs.",
		LongHelp:   `List, search and open Sourcegraph RFCs`,
		FlagSet:    rfcFlagSet,
		Exec:       rfcExec,
	}
)

func rfcExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		args = append(args, "list")
	}

	switch args[0] {
	case "list":
		return rfc.List(ctx, out)

	case "search":
		if len(args) != 2 {
			return errors.New("no search query given")
		}

		return rfc.Search(ctx, args[1], out)

	case "open":
		if len(args) != 2 {
			return errors.New("no number given")
		}

		return rfc.Open(ctx, args[1], out)
	}

	return nil
}

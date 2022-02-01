package main

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
)

var (
	versionFlagSet = flag.NewFlagSet("sg version", flag.ExitOnError)
	versionCommand = &ffcli.Command{
		Name:       "version",
		ShortUsage: "sg version",
		ShortHelp:  "Prints the sg version",
		FlagSet:    versionFlagSet,
		Exec:       versionExec,
	}
)

func versionExec(ctx context.Context, args []string) error {
	stdout.Out.Write(BuildCommit)
	return nil
}

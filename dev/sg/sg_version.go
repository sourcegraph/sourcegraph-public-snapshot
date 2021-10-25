package main

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
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
	out.Write(BuildCommit)
	return nil
}

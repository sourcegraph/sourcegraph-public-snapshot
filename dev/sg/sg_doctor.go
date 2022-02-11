package main

import (
	"context"
	"flag"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	doctorFlagSet = flag.NewFlagSet("sg doctor", flag.ExitOnError)
	doctorCommand = &ffcli.Command{
		Name:       "doctor",
		ShortUsage: "sg doctor",
		ShortHelp:  "Runs checks to test whether system is in correct state to run Sourcegraph.",
		FlagSet:    doctorFlagSet,
		Exec:       doctorExec,
	}
)

func doctorExec(ctx context.Context, args []string) error {
	return runChecks(ctx, checks)
}

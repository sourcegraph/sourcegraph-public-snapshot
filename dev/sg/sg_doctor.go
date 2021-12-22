package main

import (
	"context"
	"flag"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
)

var (
	doctorFlagSet = flag.NewFlagSet("sg doctor", flag.ExitOnError)
	doctorCommand = &ffcli.Command{
		Name:       "doctor",
		ShortUsage: "sg doctor",
		ShortHelp:  "Run the checks defined in the sg config file.",
		LongHelp: `Run the checks defined in the sg config file to make sure your system is healthy.

See the "checks:" in the configuration file.`,
		FlagSet: doctorFlagSet,
		Exec:    doctorExec,
	}
)

func doctorExec(ctx context.Context, args []string) error {
	ok, errLine := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		stdout.Out.WriteLine(errLine)
		os.Exit(1)
	}

	var checks []run.Check
	for _, c := range globalConf.Checks {
		checks = append(checks, c)
	}
	_, err := run.Checks(ctx, globalConf.Env, checks...)
	return err
}

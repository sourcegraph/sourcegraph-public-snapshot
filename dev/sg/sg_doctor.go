package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var doctorCommand = &cli.Command{
	Name:      "doctor",
	ArgsUsage: "[...checks]",
	Usage:     "Run checks to test whether system is in correct state to run Sourcegraph",
	Category:  CategoryEnv,
	Action:    doctorExec,
}

func doctorExec(cmd *cli.Context) error {
	args := cmd.Args()
	if args.Len() == 0 {
		return runChecks(cmd.Context, checks)
	}
	checksToRun := map[string]check.CheckFunc{}
	for _, arg := range args.Slice() {
		c, ok := checks[arg]
		if !ok {
			return errors.Newf("check %q not found", arg)
		}
		checksToRun[arg] = c
	}
	return runChecks(cmd.Context, checksToRun)
}

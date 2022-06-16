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

func doctorExec(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return runChecks(ctx.Context, checks)
	}
	checksToRun := map[string]check.CheckFunc{}
	for _, arg := range args {
		c, ok := checks[arg]
		if !ok {
			return errors.Newf("check %q not found", arg)
		}
		checksToRun[arg] = c
	}
	return runChecks(ctx.Context, checksToRun)
}

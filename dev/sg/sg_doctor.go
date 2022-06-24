package main

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var doctorCommand = &cli.Command{
	Name:     "doctor",
	Usage:    "DEPRECATED - Run checks to test whether system is in correct state to run Sourcegraph",
	Category: CategoryEnv,
	Action: func(ctx *cli.Context) error {
		return errors.New("DEPRECATED: use 'sg setup -check' instead")
	},
}

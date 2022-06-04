package main

import (
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/dependencies"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var setupCommandv2 = &cli.Command{
	Name:     "setupv2",
	Usage:    "Set up your local dev environment!",
	Category: CategoryEnv,
	Action: func(cmd *cli.Context) error {
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			std.Out.WriteLine(output.Styled(output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
			return NewEmptyExitErr(1)
		}

		currentOS := runtime.GOOS
		if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
			currentOS = overridesOS
		}

		// TODO
		dependencies.NewRunner(currentOS, check.IO{
			Input:  os.Stdin,
			Output: std.Out,
		})

		return nil
	},
}

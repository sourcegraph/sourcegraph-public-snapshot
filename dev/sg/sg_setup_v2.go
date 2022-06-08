package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/dependencies"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var setupCommandV2 = &cli.Command{
	Name:     "setupv2",
	Usage:    "Set up your local dev environment!",
	Category: CategoryEnv,
	Hidden:   true, // experimental
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "check",
			Usage: "Run checks and report setup state",
		},
		&cli.BoolFlag{
			Name:  "fix",
			Usage: "Fix all checks",
		},
		&cli.BoolFlag{
			Name:  "oss",
			Usage: "Omit Sourcegraph-teammate-specific setup",
		},
	},
	Action: func(cmd *cli.Context) error {
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			std.Out.WriteLine(output.Styled(output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
			return NewEmptyExitErr(1)
		}

		currentOS := runtime.GOOS
		if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
			currentOS = overridesOS
		}

		setup := dependencies.Setup(cmd.App.Reader, std.Out, dependencies.OS(currentOS))
		setup.SetDescription(func(out *std.Output) {
			printSgSetupWelcomeScreen(out)
			out.WriteAlertf("                INFO: You can quit any time by typing ctrl-c.\n")
		})

		var inRepo bool
		if _, err := root.RepositoryRoot(); err == nil {
			inRepo = true
		} else if err != root.ErrNotInsideSourcegraph {
			return err
		}

		args := dependencies.CheckArgs{
			Teammate:            !cmd.Bool("oss"),
			InRepo:              inRepo,
			ConfigFile:          configFile,
			ConfigOverwriteFile: configOverwriteFile,
		}

		switch {
		case cmd.Bool("check"):
			return setup.Check(cmd.Context, args)

		case cmd.Bool("fix"):
			return setup.Fix(cmd.Context, args)

		default:
			// Prompt for details if flags are not set
			if !cmd.IsSet("oss") {
				std.Out.WriteNoticef("Are you a Sourcegraph teammate? (Y/n) ")
				var s string
				if _, err := fmt.Scan(&s); err != nil {
					return err
				}
				args.Teammate = s == "Y"
			}
			return setup.Interactive(cmd.Context, args)
		}
	},
}

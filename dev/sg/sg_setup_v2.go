package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/dependencies"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var setupCommandV2 = &cli.Command{
	Name:     "setupv2",
	Usage:    "Validate and set up your local dev environment!",
	Category: CategoryEnv,
	Hidden:   true, // experimental
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "check",
			Aliases: []string{"c"},
			Usage:   "Run checks and report setup state",
		},
		&cli.BoolFlag{
			Name:    "fix",
			Aliases: []string{"f"},
			Usage:   "Fix all checks",
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
		setup.AnalyticsCategory = "setup"
		setup.RenderDescription = func(out *std.Output) {
			printSgSetupWelcomeScreen(out)
			out.WriteAlertf("                INFO: You can quit any time by typing ctrl-c.\n")
		}
		setup.RunPostFixChecks = true

		args := dependencies.CheckArgs{
			Teammate:            !cmd.Bool("oss"),
			ConfigFile:          configFile,
			ConfigOverwriteFile: configOverwriteFile,
		}

		switch {
		case cmd.Bool("check"):
			err := setup.Check(cmd.Context, args)
			if err != nil {
				std.Out.WriteSuggestionf("Run 'sg setup -fix' to try and automatically fix issues!")
			}
			return err

		case cmd.Bool("fix"):
			return setup.Fix(cmd.Context, args)

		default:
			// Prompt for details if flags are not set
			if !cmd.IsSet("oss") {
				std.Out.Promptf("Are you a Sourcegraph teammate? (y/n)")
				var s string
				if _, err := fmt.Scan(&s); err != nil {
					return err
				}
				args.Teammate = s == "y"
			}
			return setup.Interactive(cmd.Context, args)
		}
	},
}

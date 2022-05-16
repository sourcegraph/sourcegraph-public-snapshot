package cliutil

import (
	"flag"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func DownTo(commandName string, factory RunnerFactory, outFactory func() *output.Output, development bool) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "db",
			Usage:    "The target `schema` to modify.",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     "target",
			Usage:    `The migration to apply. Comma-separated values are accepted.`,
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "unprivileged-only",
			Usage: `Do not apply privileged migrations.`,
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "ignore-single-dirty-log",
			Usage: `Ignore a previously failed attempt if it will be immediately retried by this operation.`,
			Value: development,
		},
	}

	action := func(cmd *cli.Context) error {
		out := outFactory()

		if cmd.NArg() != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		var (
			schemaName               = cmd.String("db")
			unprivilegedOnlyFlag     = cmd.Bool("unprivileged-only")
			ignoreSingleDirtyLogFlag = cmd.Bool("ignore-single-dirty-log")
			targets                  = cmd.StringSlice("target")
		)

		versions, err := parseTargets(targets, out)
		if err != nil {
			return err
		}

		ctx := cmd.Context
		r, err := factory(ctx, []string{schemaName})
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:     schemaName,
					Type:           runner.MigrationOperationTypeTargetedDown,
					TargetVersions: versions,
				},
			},
			UnprivilegedOnly:     unprivilegedOnlyFlag,
			IgnoreSingleDirtyLog: ignoreSingleDirtyLogFlag,
		})
	}

	return &cli.Command{
		Name:        "downto",
		UsageText:   fmt.Sprintf("%s downto -db=<schema> -target=<target>,<target>,...", commandName),
		Usage:       `Revert any applied migrations that are children of the given targets - this effectively "resets" the schema to the target version`,
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      action,
	}
}

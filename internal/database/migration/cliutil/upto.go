package cliutil

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func UpTo(commandName string, factory RunnerFactory, outFactory func() *output.Output, development bool) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "db",
			Usage:    "The target `schema` to modify.",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "target",
			Usage:    "The `migration` to apply. Comma-separated values are accepted.",
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

	exec := func(cmd *cli.Context) error {
		out := outFactory()

		if cmd.NArg() != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		var (
			schemaNameFlag           = cmd.String("db")
			unprivilegedOnlyFlag     = cmd.Bool("unprivileged-only")
			ignoreSingleDirtyLogFlag = cmd.Bool("ignore-single-dirty-log")
			targetsFlag              = cmd.String("target")
		)

		targets := strings.Split(targetsFlag, ",")

		versions := make([]int, 0, len(targets))
		for _, target := range targets {
			version, err := strconv.Atoi(target)
			if err != nil {
				return err
			}

			versions = append(versions, version)
		}

		ctx := cmd.Context
		r, err := factory(ctx, []string{schemaNameFlag})
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:     schemaNameFlag,
					Type:           runner.MigrationOperationTypeTargetedUp,
					TargetVersions: versions,
				},
			},
			UnprivilegedOnly:     unprivilegedOnlyFlag,
			IgnoreSingleDirtyLog: ignoreSingleDirtyLogFlag,
		})
	}

	return &cli.Command{
		Name:        "upto",
		UsageText:   fmt.Sprintf("%s upto -db=<schema> -target=<target>,<target>,...", commandName),
		Usage:       "Ensure a given migration has been applied - may apply dependency migrations",
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      exec,
	}
}

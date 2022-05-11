package cliutil

import (
	"flag"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Undo(commandName string, factory RunnerFactory, outFactory func() *output.Output, development bool) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "db",
			Usage:    "The target `schema` to modify.",
			Required: true,
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
			schemaNameFlag           = cmd.String("db")
			ignoreSingleDirtyLogFlag = cmd.Bool("ignore-single-dirty-log")
		)

		if schemaNameFlag == "" {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a schema via -db"))
			return flag.ErrHelp
		}

		ctx := cmd.Context
		r, err := factory(ctx, []string{schemaNameFlag})
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: schemaNameFlag,
					Type:       runner.MigrationOperationTypeRevert,
				},
			},
			IgnoreSingleDirtyLog: ignoreSingleDirtyLogFlag,
		})
	}

	return &cli.Command{
		Name:        "undo",
		UsageText:   fmt.Sprintf("%s undo -db=<schema>", commandName),
		Usage:       `Revert the last migration applied - useful in local development`,
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      action,
	}
}

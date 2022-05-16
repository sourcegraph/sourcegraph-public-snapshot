package cliutil

import (
	"flag"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Up(commandName string, factory RunnerFactory, outFactory func() *output.Output, development bool) *cli.Command {
	flags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "db",
			Usage: "The target `schema(s)` to modify. Comma-separated values are accepted. Supply \"all\" to migrate all schemas.",
			Value: cli.NewStringSlice("all"),
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
			schemaNames              = cmd.StringSlice("db")
			unprivilegedOnlyFlag     = cmd.Bool("unprivileged-only")
			ignoreSingleDirtyLogFlag = cmd.Bool("ignore-single-dirty-log")
		)

		schemaNames, err := parseSchemaNames(schemaNames, out)
		if err != nil {
			return err
		}

		operations := []runner.MigrationOperation{}
		for _, schemaName := range schemaNames {
			operations = append(operations, runner.MigrationOperation{
				SchemaName: schemaName,
				Type:       runner.MigrationOperationTypeUpgrade,
			})
		}

		ctx := cmd.Context
		r, err := factory(ctx, schemaNames)
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations:           operations,
			UnprivilegedOnly:     unprivilegedOnlyFlag,
			IgnoreSingleDirtyLog: ignoreSingleDirtyLogFlag,
		})
	}

	return &cli.Command{
		Name:        "up",
		UsageText:   fmt.Sprintf("%s up [-db=<schema>]", commandName),
		Usage:       "Apply all migrations",
		Flags:       flags,
		Action:      action,
		Description: ConstructLongHelp(),
	}
}

func parseSchemaNames(schemaNames []string, out *output.Output) ([]string, error) {
	if len(schemaNames) == 1 && schemaNames[0] == "" {
		schemaNames = nil
	}

	if len(schemaNames) == 1 && schemaNames[0] == "all" {
		schemaNames = schemas.SchemaNames
	}

	if len(schemaNames) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a schema via -db"))
		return nil, flag.ErrHelp
	}

	return schemaNames, nil
}

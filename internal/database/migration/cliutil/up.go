package cliutil

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Up(commandName string, factory RunnerFactory, outFactory func() *output.Output, development bool) *cli.Command {
	schemaNamesFlag := &cli.StringSliceFlag{
		Name:  "db",
		Usage: "The target `schema(s)` to modify. Comma-separated values are accepted. Supply \"all\" to migrate all schemas.",
		Value: cli.NewStringSlice("all"),
	}
	unprivilegedOnlyFlag := &cli.BoolFlag{
		Name:  "unprivileged-only",
		Usage: `Do not apply privileged migrations.`,
		Value: false,
	}
	ignoreSingleDirtyLogFlag := &cli.BoolFlag{
		Name:  "ignore-single-dirty-log",
		Usage: `Ignore a previously failed attempt if it will be immediately retried by this operation.`,
		Value: development,
	}

	makeOptions := func(cmd *cli.Context, schemaNames []string) runner.Options {
		operations := make([]runner.MigrationOperation, 0, len(schemaNames))
		for _, schemaName := range schemaNames {
			operations = append(operations, runner.MigrationOperation{
				SchemaName: schemaName,
				Type:       runner.MigrationOperationTypeUpgrade,
			})
		}

		return runner.Options{
			Operations:           operations,
			UnprivilegedOnly:     unprivilegedOnlyFlag.Get(cmd),
			IgnoreSingleDirtyLog: ignoreSingleDirtyLogFlag.Get(cmd),
		}
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schemaNames, err := sanitizeSchemaNames(schemaNamesFlag.Get(cmd))
		if err != nil {
			return err
		}
		if len(schemaNames) == 0 {
			return flagHelp(out, "supply a schema via -db")
		}
		r, err := setupRunner(ctx, factory, schemaNames...)
		if err != nil {
			return err
		}

		return r.Run(ctx, makeOptions(cmd, schemaNames))
	})

	return &cli.Command{
		Name:        "up",
		UsageText:   fmt.Sprintf("%s up [-db=<schema>]", commandName),
		Usage:       "Apply all migrations",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNamesFlag,
			unprivilegedOnlyFlag,
			ignoreSingleDirtyLogFlag,
		},
	}
}

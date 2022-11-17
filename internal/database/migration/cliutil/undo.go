package cliutil

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Undo(commandName string, factory RunnerFactory, outFactory OutputFactory, development bool) *cli.Command {
	schemaNameFlag := &cli.StringFlag{
		Name:     "db",
		Usage:    "The target `schema` to modify.",
		Required: true,
	}

	makeOptions := func(cmd *cli.Context) runner.Options {
		return runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: schemaNameFlag.Get(cmd),
					Type:       runner.MigrationOperationTypeRevert,
				},
			},
			IgnoreSingleDirtyLog:   development,
			IgnoreSinglePendingLog: development,
		}
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		r, err := setupRunner(ctx, factory, schemaNameFlag.Get(cmd))
		if err != nil {
			return err
		}

		return r.Run(ctx, makeOptions(cmd))
	})

	return &cli.Command{
		Name:        "undo",
		UsageText:   fmt.Sprintf("%s undo -db=<schema>", commandName),
		Usage:       `Revert the last migration applied - useful in local development`,
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNameFlag,
		},
	}
}

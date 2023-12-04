package cliutil

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Validate(commandName string, factory RunnerFactory, outFactory OutputFactory) *cli.Command {
	schemaNamesFlag := &cli.StringSliceFlag{
		Name:    "schema",
		Usage:   "The target `schema(s)` to validate. Comma-separated values are accepted. Possible values are 'frontend', 'codeintel', 'codeinsights' and 'all'.",
		Value:   cli.NewStringSlice("all"),
		Aliases: []string{"db"},
	}
	skipOutOfBandMigrationsFlag := &cli.BoolFlag{
		Name:  "skip-out-of-band-migrations",
		Usage: "Do not attempt to validate out-of-band migration status.",
		Value: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schemaNames := sanitizeSchemaNames(schemaNamesFlag.Get(cmd), out)
		if len(schemaNames) == 0 {
			return flagHelp(out, "supply a schema via -db")
		}
		r, err := setupRunner(factory, schemaNames...)
		if err != nil {
			return err
		}

		if err := r.Validate(ctx, schemaNames...); err != nil {
			return err
		}

		out.WriteLine(output.Emoji(output.EmojiSuccess, "schema okay!"))

		if !skipOutOfBandMigrationsFlag.Get(cmd) {
			db, err := store.ExtractDatabase(ctx, r)
			if err != nil {
				return err
			}

			if err := oobmigration.ValidateOutOfBandMigrationRunner(ctx, db, outOfBandMigrationRunner(db)); err != nil {
				return err
			}

			out.WriteLine(output.Emoji(output.EmojiSuccess, "oobmigrations okay!"))
		}

		return nil
	})

	return &cli.Command{
		Name:        "validate",
		Usage:       "Validate the current schema",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNamesFlag,
			skipOutOfBandMigrationsFlag,
		},
	}
}

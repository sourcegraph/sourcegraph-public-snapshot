package cliutil

import (
	"context"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/multiversion"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	oobmigrations "github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func RunOutOfBandMigrations(
	commandName string,
	runnerFactory RunnerFactory,
	outFactory OutputFactory,
	registerMigratorsWithStore func(storeFactory oobmigrations.StoreFactory) oobmigration.RegisterMigratorsFunc,
) *cli.Command {
	idsFlag := &cli.IntSliceFlag{
		Name:     "id",
		Usage:    "The target migration to run. If not supplied, all migrations are run.",
		Required: false,
	}
	applyReverseFlag := &cli.BoolFlag{
		Name:     "apply-reverse",
		Usage:    "If set, run the out of band migration in reverse.",
		Required: false,
	}
	disableAnimation := &cli.BoolFlag{
		Name:     "disable-animation",
		Usage:    "If set, progress bar animations are not displayed.",
		Required: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		r, err := runnerFactory(schemas.SchemaNames)
		if err != nil {
			return err
		}
		db, err := store.ExtractDatabase(ctx, r)
		if err != nil {
			return err
		}
		registerMigrators := registerMigratorsWithStore(store.BasestoreExtractor{Runner: r})

		if err := multiversion.RunOutOfBandMigrations(
			ctx,
			db,
			false, // dry-run
			!applyReverseFlag.Get(cmd),
			!disableAnimation.Get(cmd),
			registerMigrators,
			out,
			idsFlag.Get(cmd),
		); err != nil {
			return err
		}

		return nil
	})

	return &cli.Command{
		Name:        "run-out-of-band-migrations",
		Usage:       "Run incomplete out of band migrations.",
		Description: "",
		Action:      action,
		Flags: []cli.Flag{
			idsFlag,
			applyReverseFlag,
			disableAnimation,
		},
	}
}

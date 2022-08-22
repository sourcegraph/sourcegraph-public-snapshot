package cliutil

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func RunOutOfBandMigrations(
	logger log.Logger,
	commandName string,
	runnerFactory RunnerFactory,
	outFactory OutputFactory,
	registerMigratorsWithStore func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc,
) *cli.Command {
	idFlag := &cli.IntFlag{
		Name:     "id",
		Usage:    "The target migration to run. If not supplied, all migrations are run.",
		Required: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		r, err := runnerFactory(ctx, schemas.SchemaNames)
		if err != nil {
			return err
		}
		db, err := extractDatabase(ctx, r)
		if err != nil {
			return err
		}
		registerMigrators := registerMigratorsWithStore(basestoreExtractor{r})

		var ids []int
		if id := idFlag.Get(cmd); id != 0 {
			ids = append(ids, id)
		}
		if err := runOutOfBandMigrations(
			ctx,
			db,
			false,
			registerMigrators,
			out,
			ids,
		); err != nil {
			return err
		}

		return nil
	})

	return &cli.Command{
		Name:        "run-out-of-band-migrations",
		Usage:       "Run incomplete out of band migrations (experimental).",
		Description: "",
		Action:      action,
		Flags: []cli.Flag{
			idFlag,
		},
	}
}

func runOutOfBandMigrations(
	ctx context.Context,
	db database.DB,
	dryRun bool,
	registerMigrations oobmigration.RegisterMigratorsFunc,
	out *output.Output,
	ids []int,
) error {
	store := oobmigration.NewStoreWithDB(db)
	runner := outOfBandMigrationRunnerWithStore(store)
	if err := runner.SynchronizeMetadata(ctx); err != nil {
		return err
	}
	if err := registerMigrations(ctx, db, runner); err != nil {
		return err
	}

	if len(ids) == 0 {
		migrations, err := store.List(ctx)
		if err != nil {
			return err
		}

		for _, migration := range migrations {
			ids = append(ids, migration.ID)
		}
	}

	out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "Running out of band migrations %v", ids))

	if dryRun {
		return nil
	}

	go runner.StartPartial(ids)
	defer runner.Stop()

	for range time.NewTicker(time.Second).C {
		migrations, err := getMigrations(ctx, store, ids)
		if err != nil {
			return err
		}

		sort.Slice(migrations, func(i, j int) bool { return migrations[i].ID < migrations[j].ID })

		incomplete := migrations[:0]
		for _, m := range migrations {
			if !m.Complete() {
				incomplete = append(incomplete, m)
			}
		}
		if len(incomplete) == 0 {
			break
		}
		for _, m := range incomplete {
			out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "Out of band migration #%d is at %.2f%%", m.ID, m.Progress*100))
		}
	}

	out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Out of band migrations complete"))
	return nil
}

func getMigrations(ctx context.Context, store *oobmigration.Store, ids []int) ([]oobmigration.Migration, error) {
	if len(ids) == 0 {
		return store.List(ctx)
	}

	migrations := make([]oobmigration.Migration, 0, len(ids))
	for _, id := range ids {
		migration, ok, err := store.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.Newf("unknown migration id %d", id)
		}

		migrations = append(migrations, migration)
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].ID < migrations[j].ID })

	return migrations, nil
}

type basestoreExtractor struct {
	runner Runner
}

func (r basestoreExtractor) Store(ctx context.Context, schemaName string) (*basestore.Store, error) {
	shareableStore, err := extractDB(ctx, r.runner, schemaName)
	if err != nil {
		return nil, err
	}

	return basestore.NewWithHandle(basestore.NewHandleWithDB(shareableStore, sql.TxOptions{})), nil
}

package cliutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	oobmigrations "github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		db, err := extractDatabase(ctx, r)
		if err != nil {
			return err
		}
		registerMigrators := registerMigratorsWithStore(basestoreExtractor{r})

		if err := runOutOfBandMigrations(
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

func runOutOfBandMigrations(
	ctx context.Context,
	db database.DB,
	dryRun bool,
	up bool,
	animateProgress bool,
	registerMigrations oobmigration.RegisterMigratorsFunc,
	out *output.Output,
	ids []int,
) (err error) {
	if len(ids) != 0 {
		out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "Running out of band migrations %v", ids))
		if dryRun {
			return nil
		}
	}

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
	sort.Ints(ids)

	if dryRun {
		return nil
	}

	if err := runner.UpdateDirection(ctx, ids, !up); err != nil {
		return err
	}

	go runner.StartPartial(ids)
	defer runner.Stop()
	defer func() {
		if err == nil {
			out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Out of band migrations complete"))
		} else {
			out.WriteLine(output.Linef(output.EmojiFailure, output.StyleFailure, "Out of band migrations failed: %s", err))
		}
	}()

	updateMigrationProgress, cleanup := makeOutOfBandMigrationProgressUpdater(out, ids, animateProgress)
	defer cleanup()

	ticker := time.NewTicker(time.Second).C
	for {
		migrations, err := getMigrations(ctx, store, ids)
		if err != nil {
			return err
		}
		sort.Slice(migrations, func(i, j int) bool { return migrations[i].ID < migrations[j].ID })

		for i, m := range migrations {
			updateMigrationProgress(i, m)
		}

		complete := true
		for _, m := range migrations {
			if !m.Complete() {
				if m.ApplyReverse && m.NonDestructive {
					continue
				}

				complete = false
			}
		}
		if complete {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker:
		}
	}
}

// makeOutOfBandMigrationProgressUpdater returns a two functions: `update` should be called
// when the updates to the progress of an out-of-band migration are made and should be reflected
// in the output; and `cleanup` should be called on defer when the progress object should be
// disposed.
func makeOutOfBandMigrationProgressUpdater(out *output.Output, ids []int, animateProgress bool) (
	update func(i int, m oobmigration.Migration),
	cleanup func(),
) {
	if !animateProgress || shouldDisableProgressAnimation() {
		update = func(i int, m oobmigration.Migration) {
			out.WriteLine(output.Linef("", output.StyleReset, "Migration #%d is %.2f%% complete", m.ID, m.Progress*100))
		}
		return update, func() {}
	}

	bars := make([]output.ProgressBar, 0, len(ids))
	for _, id := range ids {
		bars = append(bars, output.ProgressBar{
			Label: fmt.Sprintf("Migration #%d", id),
			Max:   1.0,
		})
	}

	progress := out.Progress(bars, nil)
	return func(i int, m oobmigration.Migration) { progress.SetValue(i, m.Progress) }, progress.Destroy
}

// shouldDisableProgressAnimation determines if progress bars should be avoided because the log level
// will create output that interferes with a stable canvas. In effect, this adds the -disable-animation
// flag when SRC_LOG_LEVEL is info or debug.
func shouldDisableProgressAnimation() bool {
	switch log.Level(os.Getenv(log.EnvLogLevel)) {
	case log.LevelDebug:
		return true
	case log.LevelInfo:
		return true

	default:
		return false
	}
}

func getMigrations(ctx context.Context, store *oobmigration.Store, ids []int) ([]oobmigration.Migration, error) {
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

	return basestore.NewWithHandle(basestore.NewHandleWithDB(log.NoOp(), shareableStore, sql.TxOptions{})), nil
}

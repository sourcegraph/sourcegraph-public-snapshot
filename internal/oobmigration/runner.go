package oobmigration

import (
	"context"
	"errors"
	"time"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

// Runner correlates out-of-band migration records in the database with a migrator instance,
// and will run each migration that has no yet completed: either reached 100% in the forward
// direction or 0% in the reverse direction.
type Runner struct {
	store           storeIface
	tickInterval    time.Duration
	refreshInterval time.Duration
	tickClock       glock.Clock
	refreshClock    glock.Clock
	migrators       map[int]Migrator
	ctx             context.Context    // root context passed to the handler
	cancel          context.CancelFunc // cancels the root context
	finished        chan struct{}      // signals that Start has finished
}

var _ goroutine.BackgroundRoutine = &Runner{}

const (
	defaultTickInterval    = time.Second
	defaultRefreshInterval = time.Second * 30
)

func NewRunnerWithDB(db dbutil.DB) *Runner {
	return newRunner(NewStoreWithDB(dbconn.Global), defaultTickInterval, defaultRefreshInterval, glock.NewRealClock(), glock.NewRealClock())
}

func newRunner(store storeIface, tickInterval, refreshInterval time.Duration, tickClock, refreshClock glock.Clock) *Runner {
	ctx, cancel := context.WithCancel(context.Background())

	return &Runner{
		store:           store,
		tickInterval:    tickInterval,
		refreshInterval: refreshInterval,
		tickClock:       tickClock,
		refreshClock:    refreshClock,
		migrators:       map[int]Migrator{},
		ctx:             ctx,
		cancel:          cancel,
		finished:        make(chan struct{}),
	}
}

// ErrMigratorConflict occurs when multiple migrator instances are registered to the same
// out-of-band migration identifier.
var ErrMigratorConflict = errors.New("migrator already registered")

// Register correlates the given migrator with the given migration identifier. An error is
// returned if a migrator is already associated with this migration.
func (r *Runner) Register(id int, migrator Migrator) error {
	if _, ok := r.migrators[id]; ok {
		return ErrMigratorConflict
	}

	r.migrators[id] = migrator
	return nil
}

// Start runs registered migrators on a loop until they complete. This method will periodically
// re-read from the database in order to refresh its current view of the migrations.
func (r *Runner) Start() {
	defer close(r.finished)

	migrationsCh := r.listMigrations(r.ctx)
	migrations := <-migrationsCh

	for {
		select {
		case migrations = <-migrationsCh:
		case <-r.tickClock.After(r.tickInterval):
		case <-r.ctx.Done():
			return
		}

		filtered := migrations[:0]

		for i := range migrations {
			progress, err := r.runMigratorForMigration(r.ctx, migrations[i])
			if err != nil {
				log15.Error("Failed to perform migration", "id", migrations[i].ID, "error", err)
			}

			completedForward := progress == 1 && !migrations[i].ApplyReverse
			completedReverse := progress == 0 && migrations[i].ApplyReverse

			if !(completedForward || completedReverse) {
				// Only check migrations that are incomplete
				filtered = append(filtered, migrations[i])
			}
		}

		migrations = filtered
	}
}

// listMigrations returns a channel that will asynchronously receive the full list of out-of-band
// migrations that exist in the database. This channel will receive a value periodically as long
// as the given context is active.
func (r *Runner) listMigrations(ctx context.Context) <-chan []Migration {
	ch := make(chan []Migration)

	go func() {
		defer close(ch)

		for {
			migrations, err := r.store.List(ctx)
			if err != nil {
				log15.Error("Failed to list out-of-band migrations", "error", err)
			}

			select {
			case ch <- migrations:
			case <-ctx.Done():
				return
			}

			select {
			case <-r.refreshClock.After(r.refreshInterval):
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// runMigratorForMigration runs a migrator, if any, registered to the given migration. If an error
// occurs in the migration function, the error will be attached to the migration so that it can be
// surfaced to an admin. If a migration function is run, regardless of it success, the migration's
// progress will be updated.
func (r *Runner) runMigratorForMigration(ctx context.Context, migration Migration) (float64, error) {
	migrator, ok := r.migrators[migration.ID]
	if !ok {
		return migration.Progress, nil
	}

	migrationFunc := migrator.Up
	if migration.ApplyReverse {
		migrationFunc = migrator.Down
	}

	if migrationErr := migrationFunc(ctx); migrationErr != nil {
		// Migration resulted in an error. All we'll do here is add this error to the
		// migration's error message list. Unless _that_ write to the database fails,
		// we'll continue along the happy path in order to update the migration, which
		// could have made additional progress before failing.

		if err := r.store.AddError(ctx, migration.ID, migrationErr.Error()); err != nil {
			return 0, err
		}
	}

	progress, err := migrator.Progress(ctx)
	if err != nil {
		return 0, err
	}
	migration.Progress = progress

	if err := r.store.UpdateProgress(ctx, migration.ID, progress); err != nil {
		return 0, err
	}

	return progress, nil
}

// Stop will cancel the context used in Start, then blocks until Start has returned.
func (r *Runner) Stop() {
	r.cancel()
	<-r.finished
}

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

	// Launch a producer goroutine that feeds a channel periodically with the migration
	// state from the database. While a new value is not available, we'll use the data
	// we saw most recently (and will use its progress field as a write-through cache).
	migrationsCh := r.listMigrations(r.ctx)

	// Block until we list our migrations for the first time. Note that this channel will be
	// closed once the context is closed, so we don't have to do a more complex select here.
	migrations := <-migrationsCh

	// Before calling Up or Down, we want to call Progress to determine if the migration can
	// be removed immediately. Each time we re-assign the migrations variable above we'll set
	// this flag to ensure we call Progress before any other action.
	shouldCheckProgress := true

	for {
		select {
		case migrations = <-migrationsCh:
			shouldCheckProgress = true
		case <-r.tickClock.After(r.tickInterval):
		case <-r.ctx.Done():
			return
		}

		if shouldCheckProgress {
			// We just fetched these migrations - see which ones are live
			migrations = r.mapFilterMigrations(r.ctx, migrations, r.updateProgressForMigration)
			shouldCheckProgress = false
		}

		// Run the migration for this tick
		migrations = r.mapFilterMigrations(r.ctx, migrations, r.runMigratorForMigration)
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

type progressFunc func(ctx context.Context, migration *Migration) error

// mapFilterMigrations runs the given progress function on each migration in the given slice.
// The progress of each migration is updated based on the progress function's return value.
// Any migration that is now marked complete is filtered from the list. The list is filtered
// in-place and a reference to the new slice (with adjusted length) is returned.
func (r *Runner) mapFilterMigrations(ctx context.Context, migrations []Migration, fn progressFunc) []Migration {
	filtered := migrations[:0]

	for i := range migrations {
		if err := fn(ctx, &migrations[i]); err != nil {
			log15.Error("Failed migration action", "id", migrations[i].ID, "error", err)
			continue
		}

		if !migrations[i].Complete() {
			filtered = append(filtered, migrations[i])
		}
	}

	return filtered
}

func (r *Runner) runMigratorForMigration(ctx context.Context, migration *Migration) error {
	if migrator, ok := r.migrators[migration.ID]; ok {
		return r.runMigrator(ctx, migration, migrator)
	}

	return nil
}

func (r *Runner) updateProgressForMigration(ctx context.Context, migration *Migration) error {
	if migrator, ok := r.migrators[migration.ID]; ok {
		return r.updateProgress(ctx, migration, migrator)
	}

	return nil
}

// runMigrator invokes the Up or Down method on the given migrator depending on the migration
// direction. If an error occurs, it will be associated in the database with the migration record.
// Regardless of the success of the migration function, the progress function on the migrator will be
// invoked and the progress written to the database.
func (r *Runner) runMigrator(ctx context.Context, migration *Migration, migrator Migrator) error {
	migrationFunc := migrator.Up
	if migration.ApplyReverse {
		migrationFunc = migrator.Down
	}

	if migrationErr := migrationFunc(ctx); migrationErr != nil {
		// Migration resulted in an error. All we'll do here is add this error to the migration's error
		// message list. Unless _that_ write to the database fails, we'll continue along the happy path
		// in order to update the migration, which could have made additional progress before failing.

		if err := r.store.AddError(ctx, migration.ID, migrationErr.Error()); err != nil {
			return err
		}
	}

	return r.updateProgress(ctx, migration, migrator)
}

// updateProgress invokes the Progress method on the given migrator, updates the Progress field of the
// given migration record, and updates the record in the database.
func (r *Runner) updateProgress(ctx context.Context, migration *Migration, migrator Migrator) error {
	progress, err := migrator.Progress(ctx)
	if err != nil {
		return err
	}

	if err := r.store.UpdateProgress(ctx, migration.ID, progress); err != nil {
		return err
	}

	migration.Progress = progress
	return nil
}

// Stop will cancel the context used in Start, then blocks until Start has returned.
func (r *Runner) Stop() {
	r.cancel()
	<-r.finished
}

// runMigrator runs the given migrator function periodically (on each read from ticker)
// while the migration is not complete. We will periodically (on each read from migrations)
// update our current view of the migration progress and (more importantly) its direction.
func runMigrator(ctx context.Context, r *Runner, migrator Migrator, migrations <-chan Migration, tickInterval time.Duration, clock glock.Clock) {
	// Get initial migration. This channel will close when the context
	// is canceled, so we don't need to do any more complex select here.
	migration, ok := <-migrations
	if !ok {
		return
	}

	// We're just starting up - refresh our progress before migrating
	if err := r.updateProgress(ctx, &migration, migrator); err != nil {
		log15.Error("Failed migration action", "id", migration.ID, "error", err)
	}

	for {
		select {
		case migration = <-migrations:
			// Refreshed our migration state

		case <-clock.After(tickInterval):
			if !migration.Complete() {
				// Run the migration only if there's something left to do
				if err := r.runMigrator(ctx, &migration, migrator); err != nil {
					log15.Error("Failed migration action", "id", migration.ID, "error", err)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

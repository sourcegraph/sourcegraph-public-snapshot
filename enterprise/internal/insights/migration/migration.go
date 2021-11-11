package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type migrator struct {
	store *basestore.Store
	// need stores for insights/dashboards too
}

func NewMigrator(store *basestore.Store) oobmigration.Migrator {
  return &migrator{store: store}
}

func (m *migrator) Progress(ctx context.Context) (float64, error) {
	// Select the total rows and the completed rows.

	// If total is 0, return 0
	// Otherwise, return completed / total

	return 0, nil
}

func (m *migrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	allJobsQueued, countAdded, err := ensureAllJobsAreQueued(tx)
	if err != nil {
		// log error?
		// update a table with a failure, or increment a retry count?
		return err
	}
	if !allJobsQueued || countAdded > 0 {
		// If not all jobs have been queued yet, we'll try again with the next call to Up
		// The idea behind "countAdded > 0" is that if the jobs were finished queueing up on this run,
		//   we want to lock in that transaction instead of risking it being rolled back by a future error.
		//   Because it's great we made it this far!
		return nil
	}

	// If we make it here, it's because all the jobs are queued. We can now safely begin picking up
	// work and doing the migrations.

	migrationComplete, workCompleted, err := performGlobalMigration(tx)
	if !migrationComplete || workCompleted {
		// Again, if it's incomplete we'll keep trying again next time.
		// And if some were completed we exit as to lock them in.
		// Same logic for the next two.
		return nil
	}

	migrationComplete, workCompleted, err := performOrgMigration(tx)
	if !migrationComplete || workCompleted {
		return nil
	}

	migrationComplete, workCompleted, err := performUserMigration(tx)
	if !migrationComplete || workCompleted {
		return nil
	}

	return nil
}

// TODO: I don't think we need this at all, do we? What would it do? Should we use it to wipe out the jobs table to restart the migration?
func (m *migrator) Down(ctx context.Context) (err error) {
	  return nil
}

func ensureAllJobsAreQueued(tx *basestore.Store) (bool, int, error) {
	// Execute SQL query to compare total settings rows we care about with total rows in settings_migration_jobs
	// If these are equal, return nil

	// Otherwise, we need to write the jobs to their rows.

	// It looks like we can use GetSettings
	// https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/internal/insights/insights.go?L47:1
	// With a filter of "All" and a prefix of "insight"

	// We won't know what type of insight it is, based on this. That will make it harder to scan it out, but
	// I don't think that will be a huge problem.

	// So we have this group of insights, then for each one we write a row into settings_migration_jobs
	// We collect any errors and keep going

	// TBD error handling here

	return false, 0, nil
}

// Instead of 3 different functions, maybe one that takes an argument? We'll see how much they have in common. Probably a lot.
func performGlobalMigration(tx *basestore.Store) (bool, bool, error) {
	// Check if the global row is marked completed. (total_items == items_completed)
	// If so, return true, 0, nil

	// Attempt to pick up the global row, setting the lock for update.
	// If we can't pick it up because it's locked,
	// return false, false, nil

	migrationComplete, workCompleted, err := performMigrationForRow(tx, globalSettingsRow)
	if err != nil {
		return false, false, err
	}
	if !migrationComplete || workCompleted {
		return false, workCompleted, nil
	}
}

func performMigrationForRow(tx *basestore.Store, settingsRow someType) (bool, bool, error) {
	// At this point we've picked up the row and are responsible for doing the migration for everything in that settings object.
	// Here I think we can use a good deal of the code that already exists

	// Caveats: we can't wipe everything out that already exists. We'll need to match on ids and check if something exists
	// before creating it. This will also make it less of an all-or-nothing transaction. If we get errors on a few of these
	// insights we can ignore them for now.

	// Side note: maybe we want like, "total items" and "items migrated" rows on each job? Maybe that would be better than
	// "completed_at", especially since we don't care when it was completed as much. We could have that too though if it helps at all.

	// Error handling: If we're keeping track of total vs completed insights/dashboards, maybe we just need a state for retries. We can do like
	// idk, 10 retries? Call them runs even. So when a runthrough is completed it increments it. And if it gets to 10 that means something is
	// seriously wrong and needs to be looked at? That can also be reset to 0 manually if need be to retry it again later.

	// Create virtual dashboard. Some function here that, once all the items are done, it does the dashboard. Hmm the state of this needs
	// to exist too; it's not part of items_completed. Another row maybe? virtual_dashboard_completed?
}

func createVirtualDashboard(tx *basestore.Store, something something) (bool, bool, error) {

}

package migration

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type migrator struct {
	insightsDB dbutil.DB
	postgresDB dbutil.DB
	settingsMigrationJobsStore store.SettingsMigrationJobsStore
}

func NewMigrator(insightsDB dbutil.DB, postgresDB dbutil.DB) oobmigration.Migrator {
	return &migrator{insightsDB: insightsDB, postgresDB: postgresDB}
}

func (m *migrator) Progress(ctx context.Context) (float64, error) {
	// Select the total rows and the completed rows.

	// If total is 0, return 0
	// Otherwise, return completed / total

	fmt.Println("CALLING PROGRESS!!")

	return 0, nil
}

// I have questions about the transactions. We're using two completely different dbs here.
// Is the transaction just across one of them? I need to read more about this, but that will take time. :(

func (m *migrator) Up(ctx context.Context) (err error) {
	fmt.Println("CALLING UP!!")

	// Initialize stores
	m.settingsMigrationJobsStore = store.NewSettingsMigrationJobsStore(m.insightsDB)

	// tx, err := m.db.Transact(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer func() { err = tx.Done(err) }()

	allJobsQueued, err := m.EnsureAllJobsAreQueued(ctx)
	if err != nil {
		// log error?
		// update a table with a failure, or increment a retry count?
		return err
	}
	if !allJobsQueued {
		fmt.Println("Not all jobs have yet been queued. Trying again.")
		return nil
	}

	fmt.Println("All jobs queued!")

	// If we make it here, it's because all the jobs are queued. We can now safely begin picking up
	// work and doing the migrations.

	migrationComplete, workCompleted, err := m.performGlobalMigration(ctx)
	if err != nil {
		return err
	}
	if !migrationComplete || workCompleted {
		// Again, if it's incomplete we'll keep trying again next time.
		// And if some were completed we exit as to lock them in.
		// Same logic for the next two.
		return nil
	}

	// migrationComplete, workCompleted, err = performOrgMigration(tx)
	// if !migrationComplete || workCompleted {
	// 	return nil
	// }

	// migrationComplete, workCompleted, err = performUserMigration(tx)
	// if !migrationComplete || workCompleted {
	// 	return nil
	// }

	return nil
}

// TODO: I don't think we need this at all, do we? What would it do? Should we use it to wipe out the jobs table to restart the migration?
func (m *migrator) Down(ctx context.Context) (err error) {
	return nil
}

func (m *migrator) EnsureAllJobsAreQueued(ctx context.Context) (bool, error) {
	allSettings, err := insights.GetSettings(ctx, m.postgresDB, insights.All, "insight")
	if err != nil {
		return false, err
	}

	distinctSettings := getDistinctSettings(allSettings)
	jobsCount, err := m.settingsMigrationJobsStore.CountSettingsMigrationJobs(ctx)
	if err != nil {
		return false, err
	}
	if jobsCount == len(distinctSettings) {
		return true, nil
	}

	for _, settings := range distinctSettings {
		userId := settings.Subject.User
		orgId := settings.Subject.Org

		// TODO: Calculate total items from the json object
		// Let's skip this for now. Definitely do-able, maybe tedious.
		totalItems := 10

		m.settingsMigrationJobsStore.CreateSettingsMigrationJob(ctx, store.CreateSettingsMigrationJobArgs{
			UserId:     userId,
			OrgId:      orgId,
			TotalItems: totalItems,
		})
	}

	// TBD error handling here
	// What if there's some error that keeps happening that we can't recover from? I can't think of what that
	// might be, but should we try and account for it?

	return false, nil
}

// We want the latest settings object for each distinct combination of (user, org, global). We don't care who authored it.
func getDistinctSettings(allSettings []*api.Settings) []*api.Settings {
	settingsMap := make(map[string]*api.Settings)

	for _, settings := range allSettings {
		key := "_"
		if settings.Subject.User != nil {
			key += string(*settings.Subject.User)
		} else {
			key += "_"
		}
		if settings.Subject.Org != nil {
			key += string(*settings.Subject.Org)
		} else {
			key += "_"
		}

		if settingsMap[key] == nil {
			settingsMap[key] = settings
		} else {
			if settingsMap[key].CreatedAt.Before(settings.CreatedAt) {
				settingsMap[key] = settings
			}
		}
	}

	var distinctSettings []*api.Settings
	for _, value := range settingsMap {
		distinctSettings = append(distinctSettings, value)
	}

	return distinctSettings
}

// Instead of 3 different functions, maybe one that takes an argument? We'll see how much they have in common. Probably a lot.
func (m *migrator) performGlobalMigration(ctx context.Context) (bool, bool, error) {
	jobs, err := m.settingsMigrationJobsStore.GetNextSettingsMigrationJobs(ctx, store.GlobalJob)
	if err != nil {
		return false, false, err
	}
	allComplete, err := m.settingsMigrationJobsStore.IsJobTypeComplete(ctx, store.GlobalJob)
	if err != nil {
		return false, false, err
	}
	if allComplete {
		return true, false, nil
	}
	// This would mean the job was locked, but not complete
	if len(jobs) == 0 {
		return false, false, nil
	}

	migrationComplete, workCompleted, err := m.performMigrationForRow(ctx, *jobs[0])
	if err != nil {
		return false, false, err
	}
	if !migrationComplete || workCompleted {
		return false, workCompleted, nil
	}

	return true, false, nil
}

// func performOrgMigration(tx *basestore.Store) (bool, bool, error) {
// 	// If there are no org rows
// 	// return true, false, nil

// 	// Check if all org rows are marked completed. (total_items == items_completed, dashboard_created == true)
// 	// If so, return true, false, nil

// 	// Attempt to pick up batchSize org rows.
// 	// If we can't pick any up it's because they are locked
// 	// return false, false, nil

// 	// Loop over the rows that we've picked up
// 	// Keep track such that if workCompleted is ever true, we make sure to return it as true from here.

// 	// First pass: just do this for ONE row

// 	migrationComplete, workCompleted, err := performMigrationForRow(tx, globalSettingsRow)
// 	if err != nil {
// 		return false, false, err
// 	}
// 	if !migrationComplete || workCompleted {
// 		return false, workCompleted, nil
// 	}

// 	return true, false, nil
// }

// func performUserMigration(tx *basestore.Store) (bool, bool, error) {
// 	// This will probably follow the same logic as orgs
// }

func (m *migrator) performMigrationForRow(ctx context.Context, settingsRow store.SettingsMigrationJob) (bool, bool, error) {
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

	return false, false, nil
}

// // Something like this? Maybe this doesn't need to be a helper function. We'll see.
// func createVirtualDashboard(tx *basestore.Store, settingsRow someType) error {
// 	// Create a dashboard for this user (or org, or global)

// 	// Fetch all of the insights for this user (or org, or global)
// 	//   Note: every insight will have exactly one grant, so this should be fine.

// 	// Then one by one attach insights to the dashboard.

// 	// If there were no errors
// 	// return nil
// }

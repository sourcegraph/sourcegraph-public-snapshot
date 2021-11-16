package migration

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type migrationBatch string

const (
	backend  migrationBatch = "backend"
	frontend migrationBatch = "frontend"
)

type migrator struct {
	insightsDB dbutil.DB
	postgresDB dbutil.DB

	settingsMigrationJobsStore *store.DBSettingsMigrationJobsStore
	settingsStore              database.SettingsStore
	insightStore               *store.InsightStore
	dashboardStore             *store.DBDashboardStore
}

func NewMigrator(insightsDB dbutil.DB, postgresDB dbutil.DB) oobmigration.Migrator {
	return &migrator{
		insightsDB:                 insightsDB,
		postgresDB:                 postgresDB,
		settingsMigrationJobsStore: store.NewSettingsMigrationJobsStore(postgresDB),
		settingsStore:              database.Settings(postgresDB),
		insightStore:               store.NewInsightStore(insightsDB),
		dashboardStore:             store.NewDashboardStore(insightsDB),
	}
}

func (m *migrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.settingsMigrationJobsStore.Query(ctx, sqlf.Sprintf(`
		SELECT CASE c2.count
				   WHEN 0 THEN 1
				   ELSE
					   CAST(c1.count AS FLOAT) / CAST(c2.count AS FLOAT) END
		FROM (SELECT COUNT(*) AS count FROM insights_settings_migration_jobs WHERE completed_at IS NOT NULL) c1,
			 (SELECT COUNT(*) AS count FROM insights_settings_migration_jobs) c2;
	`)))
	return progress, err
}

// I have questions about the transactions. We're using two completely different dbs here.
// Is the transaction just across one of them? I need to read more about this, but that will take time. :(

func (m *migrator) Up(ctx context.Context) (err error) {
	fmt.Println("CALLING UP!!")

	// tx, err := m.db.Transact(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer func() { err = tx.Done(err) }()

	migrationComplete, workCompleted, err := m.PerformGlobalMigration(ctx)
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

// Instead of 3 different functions, maybe one that takes an argument? We'll see how much they have in common. Probably a lot.
func (m *migrator) PerformGlobalMigration(ctx context.Context) (bool, bool, error) {
	jobs, err := m.settingsMigrationJobsStore.GetNextSettingsMigrationJobs(ctx, store.GlobalJob)
	if err != nil {
		fmt.Println(err)
		return false, false, err
	}
	allComplete, err := m.settingsMigrationJobsStore.IsJobTypeComplete(ctx, store.GlobalJob)
	if err != nil {
		fmt.Println(err)
		return false, false, err
	}
	if allComplete {
		fmt.Println("global jobs all complete!")
		return true, false, nil
	}
	// This would mean the job was locked, but not complete
	if len(jobs) == 0 {
		fmt.Println("global jobs locked, but not complete")
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

func (m *migrator) performMigrationForRow(ctx context.Context, job store.SettingsMigrationJob) (bool, bool, error) {
	var subject api.SettingsSubject
	if job.UserId != nil {
		userId := int32(*job.UserId)
		subject = api.SettingsSubject{User: &userId}
	} else if job.OrgId != nil {
		orgId := int32(*job.OrgId)
		subject = api.SettingsSubject{User: &orgId}
	} else {
		subject = api.SettingsSubject{Site: true}
	}
	settings, err := m.settingsStore.GetLatest(ctx, subject)
	if err != nil {
		fmt.Println(err)
		return false, false, err
	}
	if settings == nil {
		// This would mean what, the org or user was deleted before we could process it?
		// I think in that case, we just skip it.
		fmt.Println("shouldn't happen while testing")
		return true, false, nil
	}

	// fmt.Println(settings)

	langStatsInsights, err := getLangStatsInsights(ctx, *settings)
	if err != nil {
		return false, false, err
	}
	frontendInsights, err := getFrontendInsights(ctx, *settings)
	if err != nil {
		return false, false, err
	}
	backendInsights, err := getBackendInsights(ctx, *settings)
	if err != nil {
		return false, false, err
	}

	fmt.Println("lang stats:", langStatsInsights)
	fmt.Println("frontend:", frontendInsights)
	fmt.Println("backend:", backendInsights)

	totalInsights := len(langStatsInsights) + len(frontendInsights) + len(backendInsights)
	fmt.Println("total insights:", totalInsights)

	// Update row with total_insights. Then compare with migrated_insights. If they're equal, continue. If not,
	// Try migrating all of these insights.

	var migratedInsightsCount int
	if totalInsights != job.MigratedInsights {
		err = m.settingsMigrationJobsStore.UpdateTotalInsights(ctx, job.UserId, job.OrgId, totalInsights)

		migratedInsightsCount += m.migrateLangStatsInsights(ctx, langStatsInsights)
		migratedInsightsCount += m.migrateInsights(ctx, frontendInsights, frontend)
		migratedInsightsCount += m.migrateInsights(ctx, backendInsights, backend)

		err = m.settingsMigrationJobsStore.UpdateMigratedInsights(ctx, job.UserId, job.OrgId, migratedInsightsCount)
		if totalInsights != migratedInsightsCount {
			fmt.Println("Insights did not finish migrating. Exit.")
			return false, false, nil
		}
	}

	dashboards, err := getDashboards(ctx, *settings)
	if err != nil {
		return false, true, err
	}
	fmt.Println("dashboards:", dashboards)
	totalDashboards := len(dashboards)
	fmt.Println("total dashboards:", totalDashboards)

	var migratedDashboardsCount int
	if totalDashboards != job.MigratedDashboards {
		err = m.settingsMigrationJobsStore.UpdateTotalDashboards(ctx, job.UserId, job.OrgId, totalDashboards)

		migratedDashboardsCount += m.migrateDashboards(ctx, dashboards)

		err = m.settingsMigrationJobsStore.UpdateMigratedDashboards(ctx, job.UserId, job.OrgId, migratedDashboardsCount)
		if totalDashboards != migratedDashboardsCount {
			fmt.Println("Dashboards did not finish migrating. Exit.")
			return false, true, nil
		}
	}

	// TODO: Create virtual dashboard here.

	// Error handling: If we're keeping track of total vs completed insights/dashboards, maybe we just need a state for retries. We can do like
	// idk, 10 retries? Call them runs even. So when a runthrough is completed it increments it. And if it gets to 10 that means something is
	// seriously wrong and needs to be looked at? That can also be reset to 0 manually if need be to retry it again later.

	// Create virtual dashboard. Some function here that, once all the items are done, it does the dashboard. Hmm the state of this needs
	// to exist too; it's not part of items_completed. Another row maybe? virtual_dashboard_completed?

	return false, false, nil
}

// migrationContext represents a context for which we are currently migrating. If we are migrating a user setting we would populate this with their
// user ID, as well as any orgs they belong to. If we are migrating an org, we would populate this with just that orgID.
type migrationContext struct {
	userId int
	orgIds []int
}

func (m *migrator) lookupUniqueId(ctx context.Context, migration migrationContext, insightId string) (string, bool, error) {
	return basestore.ScanFirstString(m.insightStore.Query(ctx, migration.ToInsightUniqueIdQuery(insightId)))
}

func (c migrationContext) ToInsightUniqueIdQuery(insightId string) *sqlf.Query {
	similarClause := sqlf.Sprintf("unique_id similar to %s", c.buildUniqueIdCondition(insightId))
	globalClause := sqlf.Sprintf("unique_id = %s", insightId)

	q := sqlf.Sprintf("select unique_id from insight_view where %s limit 1", sqlf.Join([]*sqlf.Query{similarClause, globalClause}, "OR"))

	log.Println(q.Query(sqlf.PostgresBindVar), q.Args())
	return q
}

func (c migrationContext) buildUniqueIdCondition(insightId string) string {
	var conds []string
	for _, orgId := range c.orgIds {
		conds = append(conds, fmt.Sprintf("org-%d", orgId))
	}
	if c.userId != 0 {
		conds = append(conds, fmt.Sprintf("user-%d", c.userId))
	}
	return fmt.Sprintf("%s-%%(%s)%%", insightId, strings.Join(conds, "|"))
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

// Okay so we're going to scan through each JSON blob 4 times. One for each of the 3 insight types, and once for dashboards.

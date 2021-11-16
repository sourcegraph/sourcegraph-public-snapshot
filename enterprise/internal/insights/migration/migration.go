package migration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type migrator struct {
	insightsDB dbutil.DB
	postgresDB dbutil.DB

	settingsMigrationJobsStore *store.DBSettingsMigrationJobsStore
	settingsStore              database.SettingsStore
	insightStore               store.InsightStore
	dashboardStore             store.DashboardStore
}

func NewMigrator(insightsDB dbutil.DB, postgresDB dbutil.DB) oobmigration.Migrator {
	return &migrator{
		insightsDB:                 insightsDB,
		postgresDB:                 postgresDB,
		settingsMigrationJobsStore: store.NewSettingsMigrationJobsStore(insightsDB),
		settingsStore:              database.Settings(postgresDB),
		insightStore:               *store.NewInsightStore(insightsDB),
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

		m.settingsMigrationJobsStore.CreateSettingsMigrationJob(ctx, store.CreateSettingsMigrationJobArgs{
			UserId: userId,
			OrgId:  orgId,
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

	// Okay so now we have 4 different things to do BEFORE creating the virtual dashboard.
	// Let's worry about errors later.

	// fmt.Println(settings)

	langStatsInsights, err := getLangStatsInsights(ctx, *settings)
	frontendInsights, err := getFrontendInsights(ctx, *settings)
	backendInsights, err := getBackendInsights(ctx, *settings)

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
		fmt.Println("About to migrate langstats insights")

		migratedInsightsCount += m.migrateLangStatsInsights(ctx, langStatsInsights)
		//migratedInsightsCount += m.migrateFrontendInsights(ctx, frontendInsights)
		//migratedInsightsCount += m.migrateBackendInsights(ctx, backendInsights)
		err = m.settingsMigrationJobsStore.UpdateMigratedInsights(ctx, job.UserId, job.OrgId, migratedInsightsCount)
	}

	if totalInsights != migratedInsightsCount {
		return false, false, nil
	}

	dashboards, err := getDashboards(ctx, *settings)
	fmt.Println("dashboards:", dashboards)
	totalDashboards := len(dashboards)
	fmt.Println("total dashboards:", totalDashboards)

	if totalDashboards != job.MigratedDashboards {
		err = m.settingsMigrationJobsStore.UpdateTotalDashboards(ctx, job.UserId, job.OrgId, totalInsights)
	}

	// Update row with total_dashboards. Then compare with migrated_dashboards. If they're equal, continue. If not,
	// Try migrating all of these insights.

	// Okay so yeah we can't wipe everything out. Let's migrate 4 groups these one at a time. Dashboards last.
	// Hmm, dashboards can only succeed once all the insights have worked. Maybe we actually need to keep track
	// of both insights and dashboards seprately.. like if all except 1 insight migrates we can't do the dashboards yet.
	// Right?

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

// Okay so we're going to scan through each JSON blob 4 times. One for each of the 3 insight types, and once for dashboards.

func getLangStatsInsights(ctx context.Context, settingsRow api.Settings) ([]insights.LangStatsInsight, error) {
	prefix := "codeStatsInsights."
	var raw map[string]json.RawMessage
	results := make([]insights.LangStatsInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}

	for id, body := range raw {
		var temp insights.LangStatsInsight
		temp.ID = id
		if err := json.Unmarshal(body, &temp); err != nil {
			// a deprecated schema collides with this field name, so skip any deserialization errors
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		results = append(results, temp)
	}

	return results, nil
}

func getFrontendInsights(ctx context.Context, settingsRow api.Settings) ([]insights.SearchInsight, error) {
	prefix := "searchInsights."
	var raw map[string]json.RawMessage
	results := make([]insights.SearchInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}

	for id, body := range raw {
		var temp insights.SearchInsight
		temp.ID = id
		if err := json.Unmarshal(body, &temp); err != nil {
			// a deprecated schema collides with this field name, so skip any deserialization errors
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		results = append(results, temp)
	}

	return results, nil
}

func getBackendInsights(ctx context.Context, settingsRow api.Settings) ([]insights.SearchInsight, error) {
	prefix := "insights.allrepos"
	var raw map[string]json.RawMessage
	results := make([]insights.SearchInsight, 0)

	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}

	for id, body := range raw {
		var temp insights.SearchInsight
		temp.ID = id
		if err := json.Unmarshal(body, &temp); err != nil {
			// a deprecated schema collides with this field name, so skip any deserialization errors
			continue
		}
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		results = append(results, temp)
	}

	return results, nil
}

func getDashboards(ctx context.Context, settingsRow api.Settings) ([]insights.SettingDashboard, error) {
	prefix := "insights.dashboards"

	results := make([]insights.SettingDashboard, 0)
	var raw map[string]json.RawMessage
	raw, err := insights.FilterSettingJson(settingsRow.Contents, prefix)
	if err != nil {
		return nil, err
	}
	for _, val := range raw {
		// iterate for each instance of the prefix key in the settings. This should never be len > 1, but it's technically a map.
		temp, err := unmarshalDashboard(val, settingsRow)
		if err != nil {
			continue
		}
		results = append(results, temp...)
	}

	return results, nil
}

func unmarshalDashboard(raw json.RawMessage, settingsRow api.Settings) ([]insights.SettingDashboard, error) {
	var dict map[string]json.RawMessage
	var multi error
	result := []insights.SettingDashboard{}

	if err := json.Unmarshal(raw, &dict); err != nil {
		return result, err
	}

	for id, body := range dict {
		var temp insights.SettingDashboard
		if err := json.Unmarshal(body, &temp); err != nil {
			multi = multierror.Append(multi, err)
			continue
		}
		temp.ID = id
		temp.UserID = settingsRow.Subject.User
		temp.OrgID = settingsRow.Subject.Org

		result = append(result, temp)
	}

	return result, multi
}

func (m *migrator) migrateLangStatsInsights(ctx context.Context, toMigrate []insights.LangStatsInsight) int {
	var count, skipped, errorCount int
	for _, d := range toMigrate {
		if d.ID == "" {
			// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
			skipped++
			continue
		}
		err := migrateLangStatSeries(ctx, &m.insightStore, d)
		if err != nil {
			// we can't do anything about errors, so we will just skip it and log it
			errorCount++
			log15.Error("insights migration: error while migrating insight", "error", err)
		} else {
			count++
		}
	}
	log15.Info("insights settings migration batch complete", "batch", "langStats", "count", count, "skipped", skipped, "errors", errorCount)
	return count
}

func migrateLangStatSeries(ctx context.Context, insightStore *store.InsightStore, from insights.LangStatsInsight) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

	log15.Info("insights migration: attempting to migrate insight", "unique_id", from.ID)

	view := types.InsightView{
		Title:            from.Title,
		UniqueID:         from.ID,
		OtherThreshold:   &from.OtherThreshold,
		PresentationType: types.Pie,
	}
	series := types.InsightSeries{
		SeriesID:           ksuid.New().String(),
		Repositories:       []string{from.Repository},
		SampleIntervalUnit: string(types.Month),
	}
	var grants []store.InsightViewGrant
	if from.UserID != nil {
		grants = []store.InsightViewGrant{store.UserGrant(int(*from.UserID))}
	} else if from.OrgID != nil {
		grants = []store.InsightViewGrant{store.OrgGrant(int(*from.OrgID))}
	} else {
		grants = []store.InsightViewGrant{store.GlobalGrant()}
	}

	view, err = tx.CreateView(ctx, view, grants)
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}
	series, err = tx.CreateSeries(ctx, series)
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}
	err = tx.AttachSeriesToView(ctx, series, view, types.InsightViewSeriesMetadata{})
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight unique_id: %s", from.ID)
	}

	return nil
}

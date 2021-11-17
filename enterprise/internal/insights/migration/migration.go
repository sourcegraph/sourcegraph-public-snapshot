package migration

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/insights"
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
	orgStore                   database.OrgStore
}

func NewMigrator(insightsDB dbutil.DB, postgresDB dbutil.DB) oobmigration.Migrator {
	return &migrator{
		insightsDB:                 insightsDB,
		postgresDB:                 postgresDB,
		settingsMigrationJobsStore: store.NewSettingsMigrationJobsStore(postgresDB),
		settingsStore:              database.Settings(postgresDB),
		insightStore:               store.NewInsightStore(insightsDB),
		dashboardStore:             store.NewDashboardStore(insightsDB),
		orgStore:                   database.Orgs(postgresDB),
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
	// tx, err := m.db.Transact(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer func() { err = tx.Done(err) }()

	migrationComplete, workCompleted, err := m.performBatchMigration(ctx, store.GlobalJob)
	if err != nil {
		return err
	}
	if !migrationComplete || workCompleted {
		// Again, if it's incomplete we'll keep trying again next time.
		// And if some were completed we exit as to lock them in.
		// Same logic for the next two.
		return nil
	}

	migrationComplete, workCompleted, err = m.performBatchMigration(ctx, store.OrgJob)
	if err != nil {
		return err
	}
	if !migrationComplete || workCompleted {
		return nil
	}

	migrationComplete, workCompleted, err = m.performBatchMigration(ctx, store.UserJob)
	if err != nil {
		return err
	}
	if !migrationComplete || workCompleted {
		return nil
	}

	return nil
}

// TODO: I don't think we need this at all
func (m *migrator) Down(ctx context.Context) (err error) {
	return nil
}

func (m *migrator) performBatchMigration(ctx context.Context, jobType store.SettingsMigrationJobType) (bool, bool, error) {
	jobs, err := m.settingsMigrationJobsStore.GetNextSettingsMigrationJobs(ctx, jobType)
	if err != nil {
		fmt.Println(err)
		return false, false, err
	}
	allComplete, err := m.settingsMigrationJobsStore.IsJobTypeComplete(ctx, jobType)
	if err != nil {
		fmt.Println(err)
		return false, false, err
	}
	if allComplete {
		fmt.Println("All jobs complete for type:", jobType)
		return true, false, nil
	}
	// This would mean the jobs were locked, but not complete
	if len(jobs) == 0 {
		fmt.Println("All jobs locked, but not complete for type:", jobType)
		return false, false, nil
	}

	rowsCompleted := 0
	for _, job := range jobs {
		// TODO: Not sure what to do with these. I think I made the returns too complicated. Will re-consider.
		migrationComplete, _, _ := m.performMigrationForRow(ctx, *job)
		if migrationComplete {
			rowsCompleted++
		}
	}

	if rowsCompleted == len(jobs) {
		return true, false, nil // This return statement is not it.
	} else {
		return false, true, nil
	}
}

// I don't think this needs to return an error.. we aren't going to be doing anything with it. We should just write
// out if there's an error, upgrade runs, etc.

func (m *migrator) performMigrationForRow(ctx context.Context, job store.SettingsMigrationJob) (bool, bool, error) {
	var subject api.SettingsSubject
	var migrationContext migrationContext
	var subjectName string
	orgStore := database.Orgs(m.postgresDB)

	if job.UserId != nil {
		userId := int32(*job.UserId)
		subject = api.SettingsSubject{User: &userId}

		// when this is a user setting we need to load all of the organizations the user is a member of so that we can
		// resolve insight ID collisions as if it were in a setting cascade
		orgs, err := orgStore.GetByUserID(ctx, userId)
		if err != nil {
			return false, false, errors.Wrap(err, "OrgStoreGetByUserID")
		}
		orgIds := make([]int, 0, len(orgs))
		for _, org := range orgs {
			orgIds = append(orgIds, int(org.ID))
		}
		migrationContext.userId = int(userId)
		migrationContext.orgIds = orgIds

		userStore := database.Users(m.postgresDB)
		user, err := userStore.GetByID(ctx, userId)
		if err != nil {
			return false, false, errors.Wrap(err, "UserStoreGetByID")
		}
		subjectName = replaceIfEmpty(&user.DisplayName, user.Username)
	} else if job.OrgId != nil {
		orgId := int32(*job.OrgId)
		subject = api.SettingsSubject{Org: &orgId}
		migrationContext.orgIds = []int{*job.OrgId}
		org, err := orgStore.GetByID(ctx, orgId)
		if err != nil {
			return false, false, errors.Wrap(err, "OrgStoreGetByID")
		}
		subjectName = replaceIfEmpty(org.DisplayName, org.Name)
	} else {
		subject = api.SettingsSubject{Site: true}
		// nothing to set for migration context, it will infer global based on the lack of user / orgs
		subjectName = "Global"
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
	fmt.Println("----------- Performing migration for row:", subject)

	// First, migrate the 3 types of insights
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

	// here we are constructing a total set of all of the insights defined in this specific settings block. This will help guide us
	// to understand which insights are created here, versus which are referenced from elsewhere. This will be useful for example
	// to reconstruct the special case user / org / global dashboard
	allDefinedInsightIds := make([]string, 0, len(langStatsInsights)+len(frontendInsights)+len(backendInsights))
	for _, insight := range langStatsInsights {
		allDefinedInsightIds = append(allDefinedInsightIds, insight.ID)
	}
	for _, insight := range frontendInsights {
		allDefinedInsightIds = append(allDefinedInsightIds, insight.ID)
	}
	for _, insight := range backendInsights {
		allDefinedInsightIds = append(allDefinedInsightIds, insight.ID)
	}

	fmt.Println("lang stats:", langStatsInsights)
	fmt.Println("frontend:", frontendInsights)
	fmt.Println("backend:", backendInsights)

	totalInsights := len(langStatsInsights) + len(frontendInsights) + len(backendInsights)
	fmt.Println("total insights:", totalInsights)

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

	// Then migrate the dashboards
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
	_, err = m.createSpecialCaseDashboard(ctx, subjectName, allDefinedInsightIds, migrationContext)
	if err != nil {
		return false, false, err
	}

	// TODO: Then fill in completed_at and we're done!
	// TODO: Also increment "runs"
	// TODO: And if there are errors, write those out to error_msg.

	// Error handling: If we're keeping track of total vs completed insights/dashboards, maybe we just need a state for retries. We can do like
	// idk, 10 retries? Call them runs even. So when a runthrough is completed it increments it. And if it gets to 10 that means something is
	// seriously wrong and needs to be looked at? That can also be reset to 0 manually if need be to retry it again later.

	return true, false, nil
}

func specialCaseDashboardTitle(subjectName string) string {
	format := "%s Insights"
	if subjectName == "Global" {
		return fmt.Sprintf(format, subjectName)
	}
	return fmt.Sprintf(format, fmt.Sprintf("%s's", subjectName))
}

// replaceIfEmpty will return a string where the first argument is given priority if non-empty.
func replaceIfEmpty(firstChoice *string, replacement string) string {
	if firstChoice == nil || *firstChoice == "" {
		return replacement
	}
	return *firstChoice
}

func (m *migrator) createSpecialCaseDashboard(ctx context.Context, subjectName string, insightReferences []string, migration migrationContext) (*types.Dashboard, error) {
	created, _, err := m.createDashboard(ctx, specialCaseDashboardTitle(subjectName), insightReferences, migration)
	if err != nil {
		return nil, errors.Wrap(err, "CreateSpecialCaseDashboard")
	}
	return created, nil
}

func (m *migrator) createDashboard(ctx context.Context, title string, insightReferences []string, migration migrationContext) (_ *types.Dashboard, _ []string, err error) {
	var mapped []string
	var failed []string

	tx, err := m.dashboardStore.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() { err = tx.Done(err) }()

	for _, reference := range insightReferences {
		id, exists, err := m.lookupUniqueId(ctx, migration, reference)
		if err != nil {
			return nil, nil, err
		} else if !exists {
			failed = append(failed, reference)
		}
		mapped = append(mapped, id)
	}

	var grants []store.DashboardGrant
	if migration.userId != 0 {
		grants = append(grants, store.UserDashboardGrant(migration.userId))
	} else if len(migration.orgIds) == 1 {
		grants = append(grants, store.OrgDashboardGrant(migration.orgIds[0]))
	} else {
		grants = append(grants, store.GlobalDashboardGrant())
	}
	created, err := tx.CreateDashboard(ctx, store.CreateDashboardArgs{
		Dashboard: types.Dashboard{
			Title:      title,
			InsightIDs: mapped,
			Save:       true,
		},
		Grants: grants,
		UserID: []int{migration.userId},
		OrgID:  migration.orgIds,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "CreateDashboard")
	}

	return created, failed, nil
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

func (m *migrator) migrateDashboard(ctx context.Context, from insights.SettingDashboard) (err error) {
	tx, err := m.dashboardStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Store.Done(err) }()

	log15.Info("insights migration: migrating dashboard", "settings_unique_id", from.ID)

	mc := migrationContext{}
	if from.UserID != nil {
		orgs, err := m.orgStore.GetByUserID(ctx, *from.UserID)
		if err != nil {
			return err
		}
		orgIds := make([]int, 0, len(orgs))
		for _, org := range orgs {
			orgIds = append(orgIds, int(org.ID))
		}
		mc = migrationContext{
			userId: int(*from.UserID),
			orgIds: orgIds,
		}
	} else if from.OrgID != nil {
		mc = migrationContext{
			orgIds: []int{int(*from.OrgID)},
		}
	}

	_, _, err = m.createDashboard(ctx, from.Title, from.InsightIds, mc)
	if err != nil {
		return err
	}

	return nil
}

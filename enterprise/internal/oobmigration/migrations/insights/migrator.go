package insights

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type insightsMigrator struct {
	frontendStore *basestore.Store
	insightsStore *basestore.Store
}

func NewMigrator(frontendDB, insightsDB *basestore.Store) *insightsMigrator {
	return &insightsMigrator{
		frontendStore: frontendDB,
		insightsStore: insightsDB,
	}
}

func (m *insightsMigrator) ID() int                 { return 14 }
func (m *insightsMigrator) Interval() time.Duration { return time.Second * 10 }

func (m *insightsMigrator) Progress(ctx context.Context) (float64, error) {
	if !insights.IsEnabled() {
		return 1, nil
	}

	progress, _, err := basestore.ScanFirstFloat(m.frontendStore.Query(ctx, sqlf.Sprintf(insightsMigratorProgressQuery)))
	return progress, err
}

const insightsMigratorProgressQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:Progress
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST(c1.count AS FLOAT) / CAST(c2.count AS FLOAT)
	END
FROM
	(SELECT COUNT(*) AS count FROM insights_settings_migration_jobs WHERE completed_at IS NOT NULL) c1,
	(SELECT COUNT(*) AS count FROM insights_settings_migration_jobs) c2
`

func (m *insightsMigrator) Up(ctx context.Context) (err error) {
	if !insights.IsEnabled() {
		return nil
	}

	tx, err := m.frontendStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	jobs, err := scanJobs(tx.Query(ctx, sqlf.Sprintf(insightsMigratorUpQuery, 100)))
	if err != nil || len(jobs) == 0 {
		return err
	}

	for _, job := range jobs {
		// TODO - note we might need to differentiate between tx
		// and data-level errors, otherwise we'll always cancel the
		// entire txn. The current code is a bit spaghetti about
		// which connection its using as well.
		if migrationErr := m.performMigrationForRow(ctx, tx, job); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		}
	}

	return err
}

const insightsMigratorUpQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:Up
SELECT
	user_id,
	org_id,
	migrated_insights,
	migrated_dashboards
FROM insights_settings_migration_jobs
WHERE completed_at IS NULL
ORDER BY CASE
	WHEN global IS TRUE THEN 1
	WHEN org_id IS NOT NULL THEN 2
	WHEN user_id IS NOT NULL THEN 3
END
LIMIT %s
FOR UPDATE SKIP LOCKED
`

func (m *insightsMigrator) Down(ctx context.Context) (err error) {
	return nil
}

func (m *insightsMigrator) performMigrationForRow(ctx context.Context, tx *basestore.Store, job settingsMigrationJob) (err error) {
	cond := func() *sqlf.Query {
		if job.userID != nil {
			return sqlf.Sprintf("user_id = %s", *job.userID)
		}
		if job.orgID != nil {
			return sqlf.Sprintf("org_id = %s", *job.orgID)
		}
		return sqlf.Sprintf("global IS TRUE")
	}()

	defer func() {
		tx.Exec(ctx, sqlf.Sprintf(`UPDATE insights_settings_migration_jobs SET runs = runs + 1 WHERE %s`, cond))
	}()

	userID, orgIDs, err := func() (int, []int, error) {
		if job.userID != nil {
			// when this is a user setting we need to load all of the organizations the user is a member of so that we can
			// resolve insight ID collisions as if it were in a setting cascade
			orgIDs, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(insightsMigratorPerformMigrationForRowSelectOrgsQuery, *job.userID)))
			if err != nil {
				return 0, nil, err
			}

			return *job.userID, orgIDs, nil
		}
		if job.orgID != nil {
			return 0, []int{*job.orgID}, nil
		}
		return 0, nil, nil
	}()
	if err != nil {
		return err
	}

	subjectName, settings, err := m.getSettings(ctx, tx, job.userID, job.orgID)
	if err != nil {
		return err
	}
	if len(settings) != 0 {
		dashboards, langStatsInsights, frontendInsights, backendInsights := getInsightsFromSettings(settings[0])
		totalDashboards := len(dashboards)

		// here we are constructing a total set of all of the insights defined in this specific settings block. This will help guide us
		// to understand which insights are created here, versus which are referenced from elsewhere. This will be useful for example
		// to reconstruct the special case user / org / global dashboard
		totalInsights := len(langStatsInsights) + len(frontendInsights) + len(backendInsights)
		allDefinedInsightIds := make([]string, 0, totalInsights)
		for _, insight := range langStatsInsights {
			allDefinedInsightIds = append(allDefinedInsightIds, insight.ID)
		}
		for _, insight := range frontendInsights {
			allDefinedInsightIds = append(allDefinedInsightIds, insight.ID)
		}
		for _, insight := range backendInsights {
			allDefinedInsightIds = append(allDefinedInsightIds, insight.ID)
		}
		logDuplicates(allDefinedInsightIds)

		if totalInsights != job.migratedInsights {
			if err := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insights_settings_migration_jobs SET total_insights = %s WHERE %s`, totalInsights, cond)); err != nil {
				return err
			}

			var (
				migratedInsightsCount  int
				insightMigrationErrors error
			)
			for _, f := range []func(ctx context.Context) (int, error){
				func(ctx context.Context) (int, error) { return m.migrateLangStatsInsights(ctx, langStatsInsights) },
				func(ctx context.Context) (int, error) { return m.migrateInsights(ctx, frontendInsights, "frontend") },
				func(ctx context.Context) (int, error) { return m.migrateInsights(ctx, backendInsights, "backend") },
			} {
				count, err := f(ctx)
				insightMigrationErrors = errors.Append(insightMigrationErrors, err)
				migratedInsightsCount += count
			}

			if err := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insights_settings_migration_jobs SET migrated_insights = %s WHERE %s`, migratedInsightsCount, cond)); err != nil {
				return errors.Append(insightMigrationErrors, err)
			}

			if totalInsights != migratedInsightsCount {
				return insightMigrationErrors
			}
		}

		if totalDashboards != job.migratedDashboards {
			if err := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insights_settings_migration_jobs SET total_dashboards = %s WHERE %s`, totalDashboards, cond)); err != nil {
				return err
			}

			var (
				migratedDashboardsCount  int
				dashboardMigrationErrors error
			)
			for _, f := range []func(ctx context.Context) (int, error){
				func(ctx context.Context) (int, error) { return m.migrateDashboards(ctx, dashboards, userID, orgIDs) },
			} {
				count, err := f(ctx)
				dashboardMigrationErrors = errors.Append(dashboardMigrationErrors, err)
				migratedDashboardsCount += count
			}

			if err := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insights_settings_migration_jobs SET migrated_dashboards = %s WHERE %s`, migratedDashboardsCount, cond)); err != nil {
				return err
			}

			if totalDashboards != migratedDashboardsCount {
				return dashboardMigrationErrors
			}
		}

		if err := m.createSpecialCaseDashboard(ctx, subjectName, allDefinedInsightIds, userID, orgIDs); err != nil {
			return err
		}
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorPerformMigrationForRowUpdateJobQuery, time.Now(), cond)); err != nil {
		return errors.Wrap(err, "MarkCompleted")
	}

	return nil
}

const insightsMigratorPerformMigrationForRowSelectOrgsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:performMigrationForRow
SELECT orgs.id
FROM org_members
LEFT OUTER JOIN orgs ON org_members.org_id = orgs.id
WHERE user_id = %s AND orgs.deleted_at IS NULL
`

const insightsMigratorPerformMigrationForRowUpdateJobQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:performMigrationForRow
UPDATE insights_settings_migration_jobs SET completed_at = %s WHERE %s
`

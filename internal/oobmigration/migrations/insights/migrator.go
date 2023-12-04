package insights

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type insightsMigrator struct {
	frontendStore *basestore.Store
	insightsStore *basestore.Store
	logger        log.Logger
}

func NewMigrator(frontendDB, insightsDB *basestore.Store) *insightsMigrator {
	return &insightsMigrator{
		frontendStore: frontendDB,
		insightsStore: insightsDB,
		logger:        log.Scoped("insights-migrator"),
	}
}

func (m *insightsMigrator) ID() int                 { return 14 }
func (m *insightsMigrator) Interval() time.Duration { return time.Second * 10 }

func (m *insightsMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	if !insights.IsEnabled() {
		return 1, nil
	}

	progress, _, err := basestore.ScanFirstFloat(m.frontendStore.Query(ctx, sqlf.Sprintf(insightsMigratorProgressQuery)))
	return progress, err
}

const insightsMigratorProgressQuery = `
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
		if err := m.performMigrationJob(ctx, tx, job); err != nil {
			return err
		}
	}

	return nil
}

const insightsMigratorUpQuery = `
WITH
global_jobs AS (
	SELECT id FROM insights_settings_migration_jobs
	WHERE completed_at IS NULL AND global IS TRUE
),
org_jobs AS (
	SELECT id FROM insights_settings_migration_jobs
	WHERE completed_at IS NULL AND org_id IS NOT NULL
),
user_jobs AS (
	SELECT id FROM insights_settings_migration_jobs
	WHERE completed_at IS NULL AND user_id IS NOT NULL
),
candidates AS (
	-- Select global jobs first
	SELECT id FROM global_jobs

	-- Select org jobs only if global jobs are empty
	UNION SELECT id FROM org_jobs WHERE
		NOT EXISTS (SELECT 1 FROM global_jobs)

	-- Select user jobs only if global and org jobs are empty
	UNION SELECT id FROM user_jobs WHERE
		NOT EXISTS (SELECT 1 FROM global_jobs) AND
		NOT EXISTS (SELECT 1 FROM org_jobs)
)
SELECT
	user_id,
	org_id,
	migrated_insights,
	migrated_dashboards
FROM insights_settings_migration_jobs
WHERE id IN (SELECT id FROM candidates)
LIMIT %s
FOR UPDATE SKIP LOCKED
`

func (m *insightsMigrator) Down(ctx context.Context) (err error) {
	return nil
}

func (m *insightsMigrator) performMigrationJob(ctx context.Context, tx *basestore.Store, job insightsMigrationJob) (err error) {
	defer func() {
		if err == nil {
			// Mark job as successful on non-error exit
			if execErr := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorPerformMigrationJobUpdateJobQuery, time.Now(), makeJobCondition(job))); err != nil {
				err = errors.Append(err, errors.Wrap(execErr, "failed to mark job complete"))
			}
		}
	}()

	// Extract dashboards and insights from settings
	subjectName, settings, err := m.getSettingsForJob(ctx, tx, job)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return nil
	}
	dashboards, langStatsInsights, frontendInsights, backendInsights := getInsightsFromSettings(
		settings[0],
		m.logger,
	)

	// Perform migration of insight records
	if err := m.migrateInsightsForJob(ctx, tx, job, langStatsInsights, frontendInsights, backendInsights); err != nil {
		return err
	}

	// Perform migration of dashboard records
	uniqueIDSuffix, err := m.makeUniqueIDSuffix(ctx, tx, job)
	if err != nil {
		return err
	}
	if err := m.migrateDashboardsForJob(ctx, tx, job, dashboards, uniqueIDSuffix); err != nil {
		return err
	}

	allInsightsIDs, duplicates := extractIDsFromInsights(langStatsInsights, frontendInsights, backendInsights)
	for _, id := range duplicates {
		m.logger.Warn("duplicate insight", log.String("id", id))
	}

	if err := m.createSpecialCaseDashboard(ctx, job, subjectName, allInsightsIDs, uniqueIDSuffix); err != nil {
		return err
	}

	return nil
}

const insightsMigratorPerformMigrationJobUpdateJobQuery = `
UPDATE insights_settings_migration_jobs SET completed_at = %s WHERE %s
`

func (m *insightsMigrator) migrateInsightsForJob(
	ctx context.Context,
	tx *basestore.Store,
	job insightsMigrationJob,
	langStatsInsights []langStatsInsight,
	frontendInsights []searchInsight,
	backendInsights []searchInsight,
) error {
	totalInsights := len(langStatsInsights) + len(frontendInsights) + len(backendInsights)
	if totalInsights == job.migratedInsights {
		// Job done
		return nil
	}

	return m.invokeWithProgress(ctx, tx, job, "insights", totalInsights, []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) { return m.migrateLanguageStatsInsights(ctx, langStatsInsights) },
		func(ctx context.Context) (int, error) { return m.migrateInsights(ctx, frontendInsights, "frontend") },
		func(ctx context.Context) (int, error) { return m.migrateInsights(ctx, backendInsights, "backend") },
	})
}

func (m *insightsMigrator) migrateDashboardsForJob(
	ctx context.Context,
	tx *basestore.Store,
	job insightsMigrationJob,
	dashboards []settingDashboard,
	uniqueIDSuffix string,
) error {
	totalDashboards := len(dashboards)
	if totalDashboards == job.migratedDashboards {
		// Job done
		return nil
	}

	return m.invokeWithProgress(ctx, tx, job, "dashboards", totalDashboards, []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) {
			return m.migrateDashboards(ctx, job, dashboards, uniqueIDSuffix)
		},
	})
}

func (m *insightsMigrator) invokeWithProgress(
	ctx context.Context,
	tx *basestore.Store,
	job insightsMigrationJob,
	columnSuffix string,
	total int,
	fs []func(ctx context.Context) (int, error),
) error {
	suffix := sqlf.Sprintf(columnSuffix)
	if err := tx.Exec(ctx, sqlf.Sprintf(
		insightsMigratorInvokeWithProgressUpdateTotalQuery,
		suffix,
		total,
		makeJobCondition(job),
	)); err != nil {
		return err
	}

	var (
		migrationCount int
		migrationErr   error
	)
	for _, f := range fs {
		n, err := f(ctx)
		migrationCount += n
		migrationErr = errors.Append(migrationErr, err)
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(
		insightsMigratorInvokeWithProgressUpdateMigratedQuery,
		suffix,
		migrationCount,
		makeJobCondition(job),
	)); err != nil {
		return err
	}

	if migrationErr != nil {
		m.logger.Error("failed to migrate insights", log.Error(migrationErr))
	}

	return nil
}

const insightsMigratorInvokeWithProgressUpdateTotalQuery = `
UPDATE insights_settings_migration_jobs SET total_%s = %s WHERE %s
`

const insightsMigratorInvokeWithProgressUpdateMigratedQuery = `
UPDATE insights_settings_migration_jobs SET migrated_%s = %s WHERE %s
`

func (m *insightsMigrator) createSpecialCaseDashboard(
	ctx context.Context,
	job insightsMigrationJob,
	subjectName string,
	insightIDs []string,
	uniqueIDSuffix string,
) (err error) {
	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	return m.createDashboard(ctx, tx, job, specialCaseDashboardName(subjectName), insightIDs, uniqueIDSuffix)
}

func (m *insightsMigrator) makeUniqueIDSuffix(ctx context.Context, tx *basestore.Store, job insightsMigrationJob) (string, error) {
	userID, orgIDs, err := func() (int, []int, error) {
		if job.userID != nil {
			orgIDs, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(insightsMigratorPerformMigrationJobSelecMakeUniquIDSuffixQuery, *job.userID)))
			if err != nil {
				return 0, nil, errors.Wrap(err, "failed to select user orgs")
			}

			return int(*job.userID), orgIDs, nil
		}
		if job.orgID != nil {
			return 0, []int{int(*job.orgID)}, nil
		}
		return 0, nil, nil
	}()
	if err != nil {
		return "", err
	}

	targetsUniqueIDs := make([]string, 0, len(orgIDs)+1)
	if userID != 0 {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("user-%d", userID))
	}
	for _, orgID := range orgIDs {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("org-%d", orgID))
	}

	return "%(" + strings.Join(targetsUniqueIDs, "|") + ")%", nil
}

const insightsMigratorPerformMigrationJobSelecMakeUniquIDSuffixQuery = `
SELECT orgs.id
FROM org_members
LEFT OUTER JOIN orgs ON org_members.org_id = orgs.id
WHERE user_id = %s AND orgs.deleted_at IS NULL
`

func specialCaseDashboardName(subjectName string) string {
	if subjectName != "Global" {
		subjectName += "'s"
	}

	return fmt.Sprintf("%s Insights", subjectName)
}

func makeJobCondition(job insightsMigrationJob) *sqlf.Query {
	if job.userID != nil {
		return sqlf.Sprintf("user_id = %s", *job.userID)
	}

	if job.orgID != nil {
		return sqlf.Sprintf("org_id = %s", *job.orgID)
	}

	return sqlf.Sprintf("global IS TRUE")
}

func extractIDsFromInsights(
	langStatsInsights []langStatsInsight,
	frontendInsights []searchInsight,
	backendInsights []searchInsight,
) ([]string, []string) {
	n := len(langStatsInsights) + len(frontendInsights) + len(backendInsights)
	idMap := make(map[string]struct{}, n)
	duplicateMap := make(map[string]struct{}, n)

	add := func(id string) {
		if _, ok := idMap[id]; ok {
			duplicateMap[id] = struct{}{}
		}
		idMap[id] = struct{}{}
	}

	for _, insight := range langStatsInsights {
		add(insight.ID)
	}
	for _, insight := range frontendInsights {
		add(insight.ID)
	}
	for _, insight := range backendInsights {
		add(insight.ID)
	}

	ids := make([]string, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	duplicates := make([]string, 0, len(duplicateMap))
	for id := range duplicateMap {
		duplicates = append(duplicates, id)
	}
	sort.Strings(duplicates)

	return ids, duplicates
}

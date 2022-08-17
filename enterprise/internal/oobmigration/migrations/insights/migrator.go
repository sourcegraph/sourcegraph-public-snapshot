package insights

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"

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
	CASE
		WHEN global IS NULL THEN FALSE
		ELSE TRUE
	END AS global,
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

//
// Single migration

const schemaErrorPrefix = "insights oob migration schema error"

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

	subjectName, settings, err := func() (string, []settings, error) {
		if job.userID != nil {
			return m.getSettingsForUser(ctx, tx, *job.userID)
		}
		if job.orgID != nil {
			return m.getSettingsForOrg(ctx, tx, *job.orgID)
		}
		return m.getGlobalSettings(ctx, tx)
	}()
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

//
// Lang stat insights migration

func (m *insightsMigrator) migrateLangStatsInsights(ctx context.Context, insights []langStatsInsight) (count int, err error) {
	for _, insight := range insights {
		if migrationErr := m.migrateLangStatsInsight(ctx, insight); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrator) migrateLangStatsInsight(ctx context.Context, insight langStatsInsight) (err error) {
	if insight.ID == "" {
		// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
		// since it can never be migrated, we count it towards the total
		log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromLangStatsInsight(insight), "error msg", "insight failed to migrate due to missing id")
		return nil
	}

	numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateLangStatsInsightQuery, insight.ID)))
	if err != nil || numInsights > 0 {
		return err
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
	INSERT INTO insight_view (
		title,
		description,
		unique_id,
		default_filter_include_repo_regex,
		default_filter_exclude_repo_regex,
		default_filter_search_contexts,
		other_threshold,
		presentation_type,
	)
	VALUES (%s, %s, %s, %s, %s, %s, %s, 'PIE')
	RETURNING id
	`,
		insight.Title,
		nil, // TODO - nil ok?
		insight.ID,
		nil, // TODO - nil ok?
		nil, // TOOD - nil ok?
		pq.Array(nil),
		&insight.OtherThreshold,
	)))
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight view, unique_id: %s", insight.ID)
	}

	grantValues := func() []any {
		if insight.UserID != nil {
			return []any{viewID, *insight.UserID, nil, nil}
		}
		if insight.OrgID != nil {
			return []any{viewID, nil, *insight.OrgID, nil}
		}
		return []any{viewID, nil, nil, true}
	}()
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratormigrateSeriesInsertViewGrantQuery, grantValues...)); err != nil {
		return err
	}

	now := time.Now()
	xSeriesID := ksuid.New().String()
	seriesID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
			INSERT INTO insight_series (
				series_id,
				query,
				created_at,
				oldest_historical_at,
				last_recorded_at,
				next_recording_after,
				last_snapshot_at,
				next_snapshot_after
				repositories,
				sample_interval_unit,
				sample_interval_value,
				generated_from_capture_groups,
				just_in_time,
				generation_method,
				group_by,
				needs_migration,
			)
			VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, 'MONTH', %s, %s, %s, 'language-stats', %s, false)
			RETURNING id
		`,
		xSeriesID,
		nil, // TOOD - nil ok?
		now,
		now.Add(-time.Hour*24*7*26),
		nil, // TOOD - nil ok?
		(timeInterval{unit: "MONTH", value: 0}).StepForwards(now),
		nil, // TOOD - nil ok?
		nextSnapshot(now),
		pq.Array([]string{insight.Repository}),
		nil, // TOOD - nil ok?
		nil, // TOOD - nil ok?
		true,
		nil, // TOOD - nil ok?
	)))
	if err != nil {
		return errors.Wrapf(err, "unable to migrate insight series, unique_id: %s", insight.ID)
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
		VALUES (%s, %s, '', '')
	`,
		seriesID,
		viewID,
	)); err != nil {
		return err
	}
	// Enable the series in case it had previously been soft-deleted.
	if err := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insight_series SET deleted_at = NULL WHERE series_id = %s`, xSeriesID)); err != nil {
		return err
	}

	return nil
}

const insightsMigratorMigrateLangStatsInsightQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:migrateLangStatsInsight
SELECT COUNT(*)
FROM (SELECT * FROM insight_view WHERE unique_id = %s ORDER BY unique_id) iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE i.deleted_at IS NULL
`

//
// Insight migration

func (m *insightsMigrator) migrateInsights(ctx context.Context, insights []searchInsight, batch string) (count int, err error) {
	for _, insight := range insights {
		if migrationErr := m.migrateInsight(ctx, insight, batch); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrator) migrateInsight(ctx context.Context, insight searchInsight, batch string) error {
	if insight.ID == "" {
		// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
		// skippable error
		log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromInsight(insight), "error msg", "insight failed to migrate due to missing id")
		return nil
	}

	numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*)
		FROM (SELECT * FROM insight_view WHERE unique_id = %s ORDER BY unique_id) iv
		JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
		JOIN insight_series i ON ivs.insight_series_id = i.id
		WHERE i.deleted_at IS NULL
	`, insight.ID)))
	if err != nil || numInsights > 0 {
		return err
	}

	return migrateSeries(ctx, m.insightsStore, m.frontendStore, insight, batch)
}

func migrateSeries(ctx context.Context, insightStore *basestore.Store, workerStore *basestore.Store, from searchInsight, batch string) (err error) {
	tx, err := insightStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	now := time.Now()
	dataSeries := make([]insightSeries, len(from.Series))
	metadata := make([]insightViewSeriesMetadata, len(from.Series))

	for i, timeSeries := range from.Series {
		temp := insightSeries{
			query:              timeSeries.Query,
			createdAt:          now,
			oldestHistoricalAt: now.Add(-time.Hour * 24 * 7 * 26),
		}

		if batch == "frontend" {
			temp.repositories = from.Repositories
			if temp.repositories == nil {
				// this shouldn't be possible, but if for some reason we get here there is a malformed schema
				// we can't do anything to fix this, so skip this insight
				log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromInsight(from), "error msg", "insight failed to migrate due to missing repositories")
				return nil
			}
			interval := parseTimeInterval(from)
			temp.sampleIntervalUnit = string(interval.unit)
			temp.sampleIntervalValue = interval.value
			temp.seriesID = ksuid.New().String() // this will cause some orphan records, but we can't use the query to match because of repo / time scope. We will purge orphan records at the end of this job.
			temp.justInTime = true
			temp.generationMethod = "SEARCH"
			temp.nextSnapshotAfter = nextSnapshot(now)
			temp.nextRecordingAfter = interval.StepForwards(now)
		} else if batch == "backend" {
			temp.sampleIntervalUnit = "MONTH"
			temp.sampleIntervalValue = 1
			temp.nextRecordingAfter = nextRecording(now)
			temp.nextSnapshotAfter = nextSnapshot(now)
			temp.seriesID = ksuid.New().String()
			temp.justInTime = false
			temp.generationMethod = "SEARCH"
		}

		var series insightSeries

		// Backend series require special consideration to re-use series
		if batch == "backend" {
			rows, err := scanSeries(tx.Query(ctx, sqlf.Sprintf(`
				SELECT
					id,
					series_id,
					query,
					created_at,
					oldest_historical_at,
					last_recorded_at,
					next_recording_after,
					last_snapshot_at,
					next_snapshot_after,
					sample_interval_unit,
					sample_interval_value,
					generated_from_capture_groups,
					just_in_time,
					generation_method,
					repositories,
					group_by
				FROM insight_series
				WHERE
					(repositories = '{}' OR repositories is NULL) AND
					query = %s AND
					sample_interval_unit = %s AND
					sample_interval_value = %s AND
					generated_from_capture_groups = %s AND
					group_by IS NULL
			`,
				temp.query,
				temp.sampleIntervalUnit,
				temp.sampleIntervalValue,
				false,
			)))
			if err != nil {
				return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.seriesID)
			}
			if len(rows) > 0 {
				// If the series already exists, we can re-use that series
				series = rows[0]
			} else {
				// If it's not a backend series, we just want to create it.
				id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
					INSERT INTO insight_series (
						series_id,
						query,
						created_at,
						oldest_historical_at,
						last_recorded_at,
						next_recording_after,
						last_snapshot_at,
						next_snapshot_after,
						repositories,
						sample_interval_unit,
						sample_interval_value,
						generated_from_capture_groups,
						just_in_time,
						generation_method,
						group_by,
						needs_migration
					)
					VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, false)
					RETURNING id
				`,
					temp.seriesID,
					temp.query,
					temp.createdAt,
					temp.oldestHistoricalAt,
					temp.lastRecordedAt,
					temp.nextRecordingAfter,
					temp.lastSnapshotAt,
					temp.nextSnapshotAfter,
					pq.Array(temp.repositories),
					temp.sampleIntervalUnit,
					temp.sampleIntervalValue,
					temp.generatedFromCaptureGroups,
					temp.justInTime,
					temp.generationMethod,
					temp.groupBy,
				)))
				if err != nil {
					return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.seriesID)
				}
				temp.id = id
				series = temp

				// Also match/replace old series_points ids with the new series id
				oldId := fmt.Sprintf("s:%s", fmt.Sprintf("%X", sha256.Sum256([]byte(timeSeries.Query))))
				countUpdated, _, silentErr := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
					WITH updated AS (
						UPDATE series_points sp
						SET series_id = %s
						WHERE series_id = %s
						RETURNING sp.series_id
					)
					SELECT count(*) FROM updated;
				`,
					temp.seriesID,
					oldId,
				)))
				if silentErr != nil {
					// If the find-replace fails, it's not a big deal. It will just need to be calcuated again.
					log15.Error("error updating series_id for series_points", "series_id", temp.seriesID, "err", silentErr)
				} else if countUpdated == 0 {
					// If find-replace doesn't match any records, we still need to backfill, so just continue
				} else {
					// If the find-replace succeeded, we can do a similar find-replace on the jobs in the queue,
					// and then stamp the backfill_queued_at on the new series.

					if err := workerStore.Exec(ctx, sqlf.Sprintf("update insights_query_runner_jobs set series_id = %s where series_id = %s", temp.seriesID, oldId)); err != nil {
						// If the find-replace fails, it's not a big deal. It will just need to be calcuated again.
						log15.Error("error updating series_id for jobs", "series_id", temp.seriesID, "err", errors.Wrap(err, "updateTimeSeriesJobReferences"))
					} else {
						if silentErr := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insight_series SET backfill_queued_at = %s WHERE id = %s`, now, series.id)); silentErr != nil {
							// If the stamp fails, skip it. It will just need to be calcuated again.
							log15.Error("error updating backfill_queued_at", "series_id", temp.seriesID, "err", silentErr)
						}
					}
				}
			}
		} else {
			// If it's not a backend series, we just want to create it.
			id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
				INSERT INTO insight_series (
					series_id,
					query,
					created_at,
					oldest_historical_at,
					last_recorded_at,
					next_recording_after,
					last_snapshot_at,
					next_snapshot_after,
					repositories,
					sample_interval_unit,
					sample_interval_value,
					generated_from_capture_groups,
					just_in_time,
					generation_method,
					group_by,
					needs_migration
					)
				VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, false)
				RETURNING id
			`,
				temp.seriesID,
				temp.query,
				temp.createdAt,
				temp.oldestHistoricalAt,
				temp.lastRecordedAt,
				temp.nextRecordingAfter,
				temp.lastSnapshotAt,
				temp.nextSnapshotAfter,
				pq.Array(temp.repositories),
				temp.sampleIntervalUnit,
				temp.sampleIntervalValue,
				temp.generatedFromCaptureGroups,
				temp.justInTime,
				temp.generationMethod,
				temp.groupBy,
			)))
			if err != nil {
				return errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", from.ID, temp.seriesID)
			}
			temp.id = id
			series = temp
		}
		dataSeries[i] = series

		metadata[i] = insightViewSeriesMetadata{
			label:  timeSeries.Name,
			stroke: timeSeries.Stroke,
		}
	}

	var includeRepoRegex, excludeRepoRegex *string
	if from.Filters != nil {
		includeRepoRegex = from.Filters.IncludeRepoRegexp
		excludeRepoRegex = from.Filters.ExcludeRepoRegexp
	}

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
		INSERT INTO insight_view (
			title,
			description,
			unique_id,
			default_filter_include_repo_regex,
			default_filter_exclude_repo_regex,
			presentation_type,
		)
		VALUES (%s, %s, %s, %s, %s, %s)
		RETURNING id
	`,
		from.Title,
		from.Description,
		from.ID,
		includeRepoRegex,
		excludeRepoRegex,
		"LINE",
	)))

	grantValues := func() []any {
		if from.UserID != nil {
			return []any{viewID, *from.UserID, nil, nil}
		}
		if from.OrgID != nil {
			return []any{viewID, nil, *from.OrgID, nil}
		}
		return []any{viewID, nil, nil, true}
	}()
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratormigrateSeriesInsertViewGrantQuery, grantValues...)); err != nil {
		return err
	}

	for i, insightSeries := range dataSeries {
		if err := tx.Exec(ctx, sqlf.Sprintf(`
			INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
			VALUES (%s, %s, %s, %s)
		`,
			insightSeries.id,
			viewID,
			metadata[i].label,
			metadata[i].stroke,
		)); err != nil {
			return err
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(`
			UPDATE insight_series SET deleted_at IS NULL WHERE series_id = %s
		`, insightSeries.seriesID)); err != nil {
			return err
		}
	}
	return nil
}

const insightsMigratormigrateSeriesInsertViewGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:migarte{,LangStat}Series
INSERT INTO insight_view_grants (dashboard_id, user_id, org_id, global) VALUES (%s, %s, %s, %s)
`

//
// Dashboard migration

func (m *insightsMigrator) migrateDashboards(ctx context.Context, dashboards []settingDashboard, userID int, orgIDs []int) (count int, err error) {
	for _, dashboard := range dashboards {
		if migrationErr := m.migrateDashboard(ctx, dashboard, userID, orgIDs); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrator) migrateDashboard(ctx context.Context, from settingDashboard, userID int, orgIDs []int) (err error) {
	if from.ID == "" {
		// we need a unique ID, and if for some reason this insight doesn't have one, it can't be migrated.
		// since it can never be migrated, we count it towards the total
		log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromDashboard(from), "error msg", "dashboard failed to migrate due to missing id")
		return nil
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	var grantsQuery *sqlf.Query
	if from.UserID != nil {
		grantsQuery = sqlf.Sprintf("dg.user_id = %s", *from.UserID)
	} else if from.OrgID != nil {
		grantsQuery = sqlf.Sprintf("dg.org_id = %s", *from.OrgID)
	} else {
		grantsQuery = sqlf.Sprintf("dg.global IS TRUE")
	}

	count, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*) from dashboard
		JOIN dashboard_grants dg ON dashboard.id = dg.dashboard_id
		WHERE
			dashboard.title = %s AND
			%s
	`,
		from.Title,
		grantsQuery,
	)))
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}

	if err := m.createDashboard(ctx, tx, from.Title, from.InsightIds, userID, orgIDs); err != nil {
		return err
	}

	return nil
}

//
// Dashboard creation

func (m *insightsMigrator) createSpecialCaseDashboard(ctx context.Context, subjectName string, insightReferences []string, userID int, orgIDs []int) error {
	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if subjectName != "Global" {
		subjectName += "'s"
	}

	if err := m.createDashboard(ctx, tx, fmt.Sprintf("%s Insights", subjectName), insightReferences, userID, orgIDs); err != nil {
		return errors.Wrap(err, "CreateSpecialCaseDashboard")
	}
	return nil
}

func (m *insightsMigrator) createDashboard(ctx context.Context, tx *basestore.Store, title string, insightReferences []string, userID int, orgIDs []int) (err error) {
	targetsUniqueIDs := make([]string, 0, len(orgIDs)+1)
	if userID != 0 {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("user-%d", userID))
	}
	for _, orgID := range orgIDs {
		targetsUniqueIDs = append(targetsUniqueIDs, fmt.Sprintf("org-%d", orgID))
	}

	uniqueIDs := make([]string, 0, len(insightReferences))
	for _, reference := range insightReferences {
		id, _, err := basestore.ScanFirstString(m.insightsStore.Query(ctx, sqlf.Sprintf(
			insightsMigratorCreateDashboardSelectQuery,
			reference,
			fmt.Sprintf("%s-%%(%s)%%", reference, strings.Join(targetsUniqueIDs, "|")),
		)))
		if err != nil {
			return err
		}
		uniqueIDs = append(uniqueIDs, id)
	}

	dashboardID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorCreateDashboardInsertQuery,
		title,
	)))
	if err != nil {
		return err
	}

	indexedViewIDs := make([]*sqlf.Query, 0, len(uniqueIDs))
	for i, viewID := range uniqueIDs {
		indexedViewIDs = append(indexedViewIDs, sqlf.Sprintf(
			"(%s, %s)",
			viewID,
			fmt.Sprintf("%d", i),
		))
	}
	if len(indexedViewIDs) > 0 {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			insightsMigratorCreateDashboardInsertInsightViewQuery,
			dashboardID,
			sqlf.Join(indexedViewIDs, ", "),
			pq.Array(uniqueIDs),
		)); err != nil {
			return errors.Wrap(err, "AddViewsToDashboard")
		}
	}

	grantValues := func() []any {
		if userID != 0 {
			return []any{dashboardID, userID, nil, nil}
		}
		if len(orgIDs) != 0 {
			return []any{dashboardID, nil, orgIDs[0], nil}
		}
		return []any{dashboardID, nil, nil, true}
	}()
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorCreateDashboardInsertGrantQuery, grantValues...)); err != nil {
		return errors.Wrap(err, "AddDashboardGrants")
	}

	return nil
}

const insightsMigratorCreateDashboardSelectQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:createDashboard
SELECT
	unique_id
FROM insight_view
WHERE
	unique_id = %s OR
	unique_id SIMILAR TO %s
LIMIT 1
`

const insightsMigratorCreateDashboardInsertQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:createDashboard
INSERT INTO dashboard (title, save, type)
VALUES (%s, true, 'standard')
RETURNING id
`

const insightsMigratorCreateDashboardInsertInsightViewQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:createDashboard
INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id)
SELECT
	%s AS dashboard_id,
	insight_view.id AS insight_view_id
FROM insight_view
JOIN (VALUES %s) AS ids (id, ordering) ON ids.id = insight_view.unique_id
WHERE unique_id = ANY(%s)
ORDER BY ids.ordering
ON CONFLICT DO NOTHING
`

const insightsMigratorCreateDashboardInsertGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/migration.go:createDashboard
INSERT INTO dashboard_grants (dashboard_id, user_id, org_id, global) VALUES (%s, %s, %s, %s)
`

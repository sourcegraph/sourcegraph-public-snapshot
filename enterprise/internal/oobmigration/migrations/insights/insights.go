package insights

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// migrateInsights runs migrateInsight over each of the given values. The number of successful migrations
// are returned, along with a list of errors that occurred on failing migrations. Each migration is ran in
// a fresh transaction so that failures do not influence one another.
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
		// Soft-fail this record
		m.logger.Warn("missing insight identifier", log.String("owner", getOwnerName(insight.UserID, insight.OrgID)))
		return nil
	}
	if insight.Repositories == nil && batch == "frontend" {
		// soft-fail this record
		m.logger.Error("missing insight repositories", log.String("owner", getOwnerName(insight.UserID, insight.OrgID)))
		return nil
	}

	if numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateInsightsQuery, insight.ID))); err != nil {
		return errors.Wrap(err, "failed to count insight views")
	} else if numInsights > 0 {
		// Already migrated
		return nil
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	var (
		now                                = time.Now()
		seriesWithMetadata                 = make([]insightSeriesWithMetadata, 0, len(insight.Series))
		includeRepoRegex, excludeRepoRegex *string
	)
	if insight.Filters != nil {
		includeRepoRegex = insight.Filters.IncludeRepoRegexp
		excludeRepoRegex = insight.Filters.ExcludeRepoRegexp
	}

	for _, timeSeries := range insight.Series {
		series := insightSeries{
			seriesID:           ksuid.New().String(),
			query:              timeSeries.Query,
			createdAt:          now,
			oldestHistoricalAt: now.Add(-time.Hour * 24 * 7 * 26),
			generationMethod:   "SEARCH",
		}

		if batch == "frontend" {
			intervalUnit := parseTimeIntervalUnit(insight)
			intervalValue := parseTimeIntervalValue(insight)

			series.repositories = insight.Repositories
			series.sampleIntervalUnit = intervalUnit
			series.sampleIntervalValue = intervalValue
			series.justInTime = true
			series.nextSnapshotAfter = nextSnapshot(now)
			series.nextRecordingAfter = stepForward(now, intervalUnit, intervalValue)

		} else {
			series.sampleIntervalUnit = "MONTH"
			series.sampleIntervalValue = 1
			series.justInTime = false
			series.nextSnapshotAfter = nextSnapshot(now)
			series.nextRecordingAfter = nextRecording(now)
		}

		// Create individual insight series
		migratedSeries, err := m.migrateSeries(ctx, tx, series, timeSeries, batch, now)
		if err != nil {
			return err
		}

		seriesWithMetadata = append(seriesWithMetadata, insightSeriesWithMetadata{
			insightSeries: migratedSeries,
			label:         timeSeries.Name,
			stroke:        timeSeries.Stroke,
		})
	}

	// Create insight view record
	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateInsightInsertViewQuery,
		insight.Title,
		insight.Description,
		insight.ID,
		includeRepoRegex,
		excludeRepoRegex,
	)))
	if err != nil {
		return errors.Wrap(err, "failed to insert view")
	}

	// Create insight view series records
	for _, seriesWithMetadata := range seriesWithMetadata {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			insightsMigratorMigrateInsightInsertViewSeriesQuery,
			seriesWithMetadata.id,
			viewID,
			seriesWithMetadata.label,
			seriesWithMetadata.stroke,
		)); err != nil {
			return errors.Wrap(err, "failed to insert view series")
		}
	}

	// Create the insight view grant records
	grantArgs := append([]any{viewID}, grantTiple(insight.UserID, insight.OrgID)...)
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateInsightInsertViewGrantQuery, grantArgs...)); err != nil {
		return errors.Wrap(err, "failed to insert view grants")
	}

	return nil
}

const insightsMigratorMigrateInsightsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
SELECT COUNT(*)
FROM (
	SELECT *
	FROM insight_view
	WHERE unique_id = %s
	ORDER BY unique_id
) iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE i.deleted_at IS NULL
`

const insightsMigratorMigrateInsightInsertViewQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
INSERT INTO insight_view (
	title,
	description,
	unique_id,
	default_filter_include_repo_regex,
	default_filter_exclude_repo_regex,
	presentation_type
)
VALUES (%s, %s, %s, %s, %s, 'LINE')
RETURNING id
`

const insightsMigratorMigrateInsightInsertViewSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
VALUES (%s, %s, %s, %s)
`

const insightsMigratorMigrateInsightInsertViewGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
INSERT INTO insight_view_grants (insight_view_id, user_id, org_id, global)
VALUES (%s, %s, %s, %s)
`

func (m *insightsMigrator) migrateSeries(ctx context.Context, tx *basestore.Store, series insightSeries, timeSeries timeSeries, batch string, now time.Time) (insightSeries, error) {
	series, err := m.getOrCreateSeries(ctx, tx, series)
	if err != nil {
		return insightSeries{}, err
	}

	if batch == "backend" {
		if err := m.migrateBackendSeries(ctx, tx, series, timeSeries, now); err != nil {
			return insightSeries{}, err
		}
	}

	return series, nil
}

func (m *insightsMigrator) getOrCreateSeries(ctx context.Context, tx *basestore.Store, series insightSeries) (insightSeries, error) {
	if existingSeries, ok, err := scanFirstSeries(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorGetOrCreateSeriesSelectSeriesQuery,
		series.query,
		series.sampleIntervalUnit,
		series.sampleIntervalValue,
		false,
	))); err != nil {
		return insightSeries{}, errors.Wrap(err, "failed to select series")
	} else if ok {
		// Re-use existing series
		return existingSeries, nil
	}

	return m.createSeries(ctx, tx, series)
}

const insightsMigratorGetOrCreateSeriesSelectSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:getOrCreateSeries
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
`

func (m *insightsMigrator) createSeries(ctx context.Context, tx *basestore.Store, series insightSeries) (insightSeries, error) {
	id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorCreateSeriesQuery,
		series.seriesID,
		series.query,
		series.createdAt,
		series.oldestHistoricalAt,
		series.lastRecordedAt,
		series.nextRecordingAfter,
		series.lastSnapshotAt,
		series.nextSnapshotAfter,
		pq.Array(series.repositories),
		series.sampleIntervalUnit,
		series.sampleIntervalValue,
		series.generatedFromCaptureGroups,
		series.justInTime,
		series.generationMethod,
		series.groupBy,
	)))
	if err != nil {
		return insightSeries{}, errors.Wrapf(err, "failed to insert series")
	}

	series.id = id
	return series, nil
}

const insightsMigratorCreateSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:createSeries
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
`

func (m *insightsMigrator) migrateBackendSeries(ctx context.Context, tx *basestore.Store, series insightSeries, timeSeries timeSeries, now time.Time) error {
	oldID := hashID(timeSeries.Query)

	// Replace old series points with new series identifier
	numPointsUpdated, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateBackendSeriesUpdateSeriesPointsQuery,
		series.seriesID,
		oldID,
	)))
	if err != nil {
		// soft-error (migration txn is preserved)
		m.logger.Error("failed to update series points", log.Error(err))
		return nil
	}
	if numPointsUpdated == 0 {
		// No records matched, continue - backfill will be required later
		return nil
	}

	// Replace old jobs with new series identifier
	if err := m.frontendStore.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateBackendSeriesUpdateJobsQuery, series.seriesID, oldID)); err != nil {
		// soft-error (migration txn is preserved)
		m.logger.Error("failed to update seriesID on insights jobs", log.Error(err))
		return nil
	}

	// Update backfill_queued_at on the new series on success
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateBackendSeriesUpdateBackfillQueuedAtQuery, now, series.id)); err != nil {
		return err
	}

	return nil
}

const insightsMigratorMigrateBackendSeriesUpdateSeriesPointsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateBackendSeries
WITH updated AS (
	UPDATE series_points sp
	SET series_id = %s
	WHERE series_id = %s
	RETURNING sp.series_id
)
SELECT count(*) FROM updated;
`

const insightsMigratorMigrateBackendSeriesUpdateJobsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateBackendSeries
UPDATE insights_query_runner_jobs SET series_id = %s WHERE series_id = %s
`

const insightsMigratorMigrateBackendSeriesUpdateBackfillQueuedAtQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateBackendSeries
UPDATE insight_series SET backfill_queued_at = %s WHERE id = %s
`

func getOwnerName(userID, orgID *int32) string {
	if userID != nil {
		return fmt.Sprintf("user id %d", *userID)
	} else if orgID != nil {
		return fmt.Sprintf("org id %d", *orgID)
	} else {
		return "global"
	}
}

func grantTiple(userID, orgID *int32) []any {
	if userID != nil {
		return []any{*userID, nil, nil}
	} else if orgID != nil {
		return []any{nil, *orgID, nil}
	} else {
		return []any{nil, nil, true}
	}
}

func hashID(query string) string {
	return fmt.Sprintf("s:%s", fmt.Sprintf("%X", sha256.Sum256([]byte(query))))
}
func nextSnapshot(current time.Time) time.Time {
	year, month, day := current.In(time.UTC).Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
}

func nextRecording(current time.Time) time.Time {
	year, month, _ := current.In(time.UTC).Date()
	return time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
}

func stepForward(now time.Time, intervalUnit string, intervalValue int) time.Time {
	switch intervalUnit {
	case "YEAR":
		return now.AddDate(intervalValue, 0, 0)
	case "MONTH":
		return now.AddDate(0, intervalValue, 0)
	case "WEEK":
		return now.AddDate(0, 0, 7*intervalValue)
	case "DAY":
		return now.AddDate(0, 0, intervalValue)
	case "HOUR":
		return now.Add(time.Hour * time.Duration(intervalValue))
	default:
		return now
	}
}

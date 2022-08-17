package insights

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

	numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateInsightsQuery, insight.ID)))
	if err != nil || numInsights > 0 {
		return errors.Wrap(err, "failed to count insight views")
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	var (
		now                                = time.Now()
		dataSeries                         = make([]insightSeries, len(insight.Series))
		metadata                           = make([]insightViewSeriesMetadata, len(insight.Series))
		includeRepoRegex, excludeRepoRegex *string
	)
	if insight.Filters != nil {
		includeRepoRegex = insight.Filters.IncludeRepoRegexp
		excludeRepoRegex = insight.Filters.ExcludeRepoRegexp
	}

	for i, timeSeries := range insight.Series {
		temp, ok := func() (insightSeries, bool) {
			temp := insightSeries{
				query:              timeSeries.Query,
				createdAt:          now,
				oldestHistoricalAt: now.Add(-time.Hour * 24 * 7 * 26),
			}

			switch batch {
			case "frontend":
				temp.repositories = insight.Repositories
				if temp.repositories == nil {
					// this shouldn't be possible, but if for some reason we get here there is a malformed schema
					// we can't do anything to fix this, so skip this insight
					log15.Error(schemaErrorPrefix, "owner", getOwnerNameFromInsight(insight), "error msg", "insight failed to migrate due to missing repositories")
					return insightSeries{}, false
				}
				interval := parseTimeInterval(insight)
				temp.sampleIntervalUnit = string(interval.unit)
				temp.sampleIntervalValue = interval.value
				temp.seriesID = ksuid.New().String() // this will cause some orphan records, but we can't use the query to match because of repo / time scope. We will purge orphan records at the end of this job.
				temp.justInTime = true
				temp.generationMethod = "SEARCH"
				temp.nextSnapshotAfter = nextSnapshot(now)
				temp.nextRecordingAfter = interval.StepForwards(now)

			case "backend":
				temp.sampleIntervalUnit = "MONTH"
				temp.sampleIntervalValue = 1
				temp.nextRecordingAfter = nextRecording(now)
				temp.nextSnapshotAfter = nextSnapshot(now)
				temp.seriesID = ksuid.New().String()
				temp.justInTime = false
				temp.generationMethod = "SEARCH"
			}

			return temp, true
		}()
		if !ok {
			return nil
		}

		series, err := m.migrateSeries(ctx, tx, temp, insight, timeSeries, batch, now)
		if err != nil {
			return err
		}
		dataSeries[i] = series

		metadata[i] = insightViewSeriesMetadata{
			label:  timeSeries.Name,
			stroke: timeSeries.Stroke,
		}
	}

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateInsightInsertViewQuery,
		insight.Title,
		insight.Description,
		insight.ID,
		includeRepoRegex,
		excludeRepoRegex,
	)))

	grantValues := func() []any {
		if insight.UserID != nil {
			return []any{viewID, *insight.UserID, nil, nil}
		}
		if insight.OrgID != nil {
			return []any{viewID, nil, *insight.OrgID, nil}
		}
		return []any{viewID, nil, nil, true}
	}()
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateInsightInsertViewGrantQuery, grantValues...)); err != nil {
		return err
	}

	for i, insightSeries := range dataSeries {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			insightsMigratorMigrateInsightInsertViewSeriesQuery,
			insightSeries.id,
			viewID,
			metadata[i].label,
			metadata[i].stroke,
		)); err != nil {
			return err
		}

		// Enable the series in case it had previously been soft-deleted
		if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateInsightEnableSeriesQuery, insightSeries.seriesID)); err != nil {
			return err
		}
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
	presentation_type,
)
VALUES (%s, %s, %s, %s, %s, 'LINE')
RETURNING id
`

const insightsMigratorMigrateInsightInsertViewGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
INSERT INTO insight_view_grants (dashboard_id, user_id, org_id, global)
VALUES (%s, %s, %s, %s)
`

const insightsMigratorMigrateInsightInsertViewSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
VALUES (%s, %s, %s, %s)
`

const insightsMigratorMigrateInsightEnableSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateInsight
UPDATE insight_series SET deleted_at IS NULL WHERE series_id = %s
`

func (m *insightsMigrator) migrateSeries(ctx context.Context, tx *basestore.Store, temp insightSeries, insight searchInsight, timeSeries timeSeries, batch string, now time.Time) (insightSeries, error) {
	if batch != "backend" {
		// If it's not a backend series, we just want to create it
		return m.createSeries(ctx, tx, temp, insight)
	}

	rows, err := scanSeries(tx.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateSeriesSelectSeriesQuery, temp.query, temp.sampleIntervalUnit, temp.sampleIntervalValue, false)))
	if err != nil {
		return insightSeries{}, errors.Wrap(err, "failed to select series")
	}
	if len(rows) > 0 {
		// re-use existing series
		return rows[0], nil
	}

	temp, err = m.createSeries(ctx, tx, temp, insight)
	if err != nil {
		return insightSeries{}, err
	}

	oldID := fmt.Sprintf("s:%s", fmt.Sprintf("%X", sha256.Sum256([]byte(timeSeries.Query))))

	// Match/replace old series_points ids with the new series id
	if numPointsUpdated, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateSeriesUpdateSeriesPointsQuery,
		temp.seriesID,
		oldID,
	))); err != nil || numPointsUpdated == 0 {
		if err != nil {
			log15.Error("error updating series_id for series_points", "series_id", temp.seriesID, "err", err)
		}
		return temp, nil
	}

	// Try to do a similar find-replace on the jobs in the queue
	if err := m.frontendStore.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateSeriesUpdateJobsQuery, temp.seriesID, oldID)); err != nil {
		log15.Error("error updating series_id for jobs", "series_id", temp.seriesID, "err", errors.Wrap(err, "updateTimeSeriesJobReferences"))
		return temp, nil
	}

	// Stamp the backfill_queued_at on the new series
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateSeriesUpdateBackfillQueuedAtQuery, now, temp.id)); err != nil {
		log15.Error("error updating backfill_queued_at", "series_id", temp.seriesID, "err", err)
		return temp, nil
	}

	return temp, nil
}

const insightsMigratorMigrateSeriesSelectSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
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

const insightsMigratorMigrateSeriesUpdateSeriesPointsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
WITH updated AS (
	UPDATE series_points sp
	SET series_id = %s
	WHERE series_id = %s
	RETURNING sp.series_id
)
SELECT count(*) FROM updated;
`

const insightsMigratorMigrateSeriesUpdateJobsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
UPDATE insights_query_runner_jobs SET series_id = %s WHERE series_id = %s
`

const insightsMigratorMigrateSeriesUpdateBackfillQueuedAtQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
UPDATE insight_series SET backfill_queued_at = %s WHERE id = %s
`

func (m *insightsMigrator) createSeries(ctx context.Context, tx *basestore.Store, temp insightSeries, insight searchInsight) (insightSeries, error) {
	id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorCreateSeriesQuery,
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
		return insightSeries{}, errors.Wrapf(err, "failed to insert series")
	}

	temp.id = id
	return temp, nil
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

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

	return m.migrateSeries(ctx, insight, batch)
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

func (m *insightsMigrator) migrateSeries(ctx context.Context, insight searchInsight, batch string) (err error) {
	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	now := time.Now()
	dataSeries := make([]insightSeries, len(insight.Series))
	metadata := make([]insightViewSeriesMetadata, len(insight.Series))

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

		series, err := func() (insightSeries, error) {
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
					return insightSeries{}, errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", insight.ID, temp.seriesID)
				}
				if len(rows) > 0 {
					// If the series already exists, we can re-use that series
					return rows[0], nil
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
						return insightSeries{}, errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", insight.ID, temp.seriesID)
					}
					temp.id = id

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

						if err := m.frontendStore.Exec(ctx, sqlf.Sprintf("update insights_query_runner_jobs set series_id = %s where series_id = %s", temp.seriesID, oldId)); err != nil {
							// If the find-replace fails, it's not a big deal. It will just need to be calcuated again.
							log15.Error("error updating series_id for jobs", "series_id", temp.seriesID, "err", errors.Wrap(err, "updateTimeSeriesJobReferences"))
						} else {
							if silentErr := tx.Exec(ctx, sqlf.Sprintf(`UPDATE insight_series SET backfill_queued_at = %s WHERE id = %s`, now, temp.id)); silentErr != nil {
								// If the stamp fails, skip it. It will just need to be calcuated again.
								log15.Error("error updating backfill_queued_at", "series_id", temp.seriesID, "err", silentErr)
							}
						}
					}

					return temp, nil
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
					return insightSeries{}, errors.Wrapf(err, "unable to migrate insight unique_id: %s series_id: %s", insight.ID, temp.seriesID)
				}
				temp.id = id
				return temp, nil
			}
		}()
		if err != nil {
			return err
		}
		dataSeries[i] = series

		metadata[i] = insightViewSeriesMetadata{
			label:  timeSeries.Name,
			stroke: timeSeries.Stroke,
		}
	}

	var includeRepoRegex, excludeRepoRegex *string
	if insight.Filters != nil {
		includeRepoRegex = insight.Filters.IncludeRepoRegexp
		excludeRepoRegex = insight.Filters.ExcludeRepoRegexp
	}

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateSeriesInsertViewQuery,
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
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateSeriesInsertViewGrantQuery, grantValues...)); err != nil {
		return err
	}

	for i, insightSeries := range dataSeries {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			insightsMigratorMigrateSeriesInsertViewSeriesQuery,
			insightSeries.id,
			viewID,
			metadata[i].label,
			metadata[i].stroke,
		)); err != nil {
			return err
		}

		// Enable the series in case it had previously been soft-deleted
		if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateSeriesQuery, insightSeries.seriesID)); err != nil {
			return err
		}
	}
	return nil
}

const insightsMigratorMigrateSeriesInsertViewQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
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

const insightsMigratorMigrateSeriesInsertViewGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
INSERT INTO insight_view_grants (dashboard_id, user_id, org_id, global)
VALUES (%s, %s, %s, %s)
`

const insightsMigratorMigrateSeriesInsertViewSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
VALUES (%s, %s, %s, %s)
`

const insightsMigratorMigrateSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/insights.go:migrateSeries
UPDATE insight_series SET deleted_at IS NULL WHERE series_id = %s
`

// const insightsMigratorMigrateSeriesQuery = `
// -- source: enterprise/internal/oobmigration/migrations/insights/migration.go:migrateSeries
// `

// const insightsMigratorMigrateSeriesQuery = `
// -- source: enterprise/internal/oobmigration/migrations/insights/migration.go:migrateSeries
// `

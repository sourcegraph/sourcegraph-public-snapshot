package insights

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

	numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateLangStatsInsightCountInsightsQuery, insight.ID)))
	if err != nil || numInsights > 0 {
		return errors.Wrap(err, "failed to count insights")
	}

	tx, err := m.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateLangStatsInsightInsertViewQuery,
		insight.Title,
		insight.ID,
		insight.OtherThreshold,
	)))
	if err != nil {
		return errors.Wrap(err, "failed to insert view")
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateLangstatsInsightInsertViewGrantQuery, append([]any{viewID}, grantValues2(insight.UserID, insight.OrgID))...)); err != nil {
		return errors.Wrap(err, "failed to insert view grant")
	}

	now := time.Now()
	seriesID := ksuid.New().String()

	insightSeriesID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateLangStatsInsightInsertSeriesQuery,
		seriesID,
		now,
		now.Add(-time.Hour*24*7*26), // 6 months
		(timeInterval{unit: "MONTH", value: 0}).StepForwards(now),
		nextSnapshot(now),
		pq.Array([]string{insight.Repository}),
	)))
	if err != nil {
		return errors.Wrap(err, "failed to insert series")
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateLangStatsInsightInsertViewSeriesQuery, insightSeriesID, viewID)); err != nil {
		return errors.Wrap(err, "failed to insert view series")
	}

	// Enable the series in case it had previously been soft-deleted
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateLangStatsInsightEnableSeriesQuery, seriesID)); err != nil {
		return errors.Wrap(err, "failed to enable series")
	}

	return nil
}

const insightsMigratorMigrateLangStatsInsightCountInsightsQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/langstat.go:migrateLangStatsInsight
SELECT COUNT(*)
FROM (SELECT * FROM insight_view WHERE unique_id = %s ORDER BY unique_id) iv
JOIN insight_view_series ivs ON iv.id = ivs.insight_view_id
JOIN insight_series i ON ivs.insight_series_id = i.id
WHERE i.deleted_at IS NULL
`

// Note: these columns were never set
//   - description
//   - default_filter_include_repo_regex
//   - default_filter_exclude_repo_regex
//   - default_filter_search_contexts
const insightsMigratorMigrateLangStatsInsightInsertViewQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/langstat.go:migrateLangStatsInsight
INSERT INTO insight_view (title, unique_id, other_threshold, presentation_type)
VALUES (%s, %s, %s, 'PIE')
RETURNING id
`

const insightsMigratorMigrateLangstatsInsightInsertViewGrantQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/langstat.go:migrateLangStatsInsight
INSERT INTO insight_view_grants (dashboard_id, user_id, org_id, global)
VALUES (%s, %s, %s, %s)
`

// Note: these columns were never set
//  - query
//  - last_recorded_at
//  - last_snapshot_at
//  - sample_interval_value
//  - generated_from_capture_groups
//  - group_by

const insightsMigratorMigrateLangStatsInsightInsertSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/langstat.go:migrateLangStatsInsight
	INSERT INTO insight_series (
		series_id,
		created_at,
		oldest_historical_at,
		next_recording_after,
		next_snapshot_after,
		repositories,
		sample_interval_unit,
		just_in_time,
		generation_method,
		needs_migration,
	)
	VALUES (%s, %s, %s, %s, %s, %s, 'MONTH', true, 'language-stats', false)
	RETURNING id
`

const insightsMigratorMigrateLangStatsInsightInsertViewSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/langstat.go:migrateLangStatsInsight
INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
VALUES (%s, %s, '', '')
`

const insightsMigratorMigrateLangStatsInsightEnableSeriesQuery = `
-- source: enterprise/internal/oobmigration/migrations/insights/langstat.go:migrateLangStatsInsight
UPDATE insight_series SET deleted_at = NULL WHERE series_id = %s
`

package insights

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// migrateLanguageStatsInsights runs migrateLanguageStatsInsight over each of the given values. The number of successful migrations
// are returned, along with a list of errors that occurred on failing migrations. Each migration is ran in a fresh transaction
// so that failures do not influence one another.
func (m *insightsMigrator) migrateLanguageStatsInsights(ctx context.Context, insights []langStatsInsight) (count int, err error) {
	for _, insight := range insights {
		if migrationErr := m.migrateLanguageStatsInsight(ctx, insight); migrationErr != nil {
			err = errors.Append(err, migrationErr)
		} else {
			count++
		}
	}

	return count, err
}

func (m *insightsMigrator) migrateLanguageStatsInsight(ctx context.Context, insight langStatsInsight) (err error) {
	if insight.ID == "" {
		// Soft-fail this record
		m.logger.Warn("missing language-stat insight identifier", log.String("owner", getOwnerName(insight.UserID, insight.OrgID)))
		return nil
	}

	if numInsights, _, err := basestore.ScanFirstInt(m.insightsStore.Query(ctx, sqlf.Sprintf(insightsMigratorMigrateLanguageStatsInsightCountInsightsQuery, insight.ID))); err != nil {
		return errors.Wrap(err, "failed to count insights")
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
		now      = time.Now()
		seriesID = ksuid.New().String()
	)

	// Create insight view record
	viewID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateLanguageStatsInsightInsertViewQuery,
		insight.Title,
		insight.ID,
		insight.OtherThreshold,
	)))
	if err != nil {
		return errors.Wrap(err, "failed to insert view")
	}

	// Create insight series record
	insightSeriesID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		insightsMigratorMigrateLanguageStatsInsightInsertSeriesQuery,
		seriesID,
		now,
		now.Add(-time.Hour*24*7*26), // 6 months
		now,
		nextSnapshot(now),
		pq.Array([]string{insight.Repository}),
	)))
	if err != nil {
		return errors.Wrap(err, "failed to insert series")
	}

	// Create insight view series record
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateLanguageStatsInsightInsertViewSeriesQuery, insightSeriesID, viewID)); err != nil {
		return errors.Wrap(err, "failed to insert view series")
	}

	// Create insight view grant records
	grantArgs := append([]any{viewID}, grantTiple(insight.UserID, insight.OrgID)...)
	if err := tx.Exec(ctx, sqlf.Sprintf(insightsMigratorMigrateLanguageStatsInsightInsertViewGrantQuery, grantArgs...)); err != nil {
		return errors.Wrap(err, "failed to insert view grant")
	}

	return nil
}

const insightsMigratorMigrateLanguageStatsInsightCountInsightsQuery = `
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
const insightsMigratorMigrateLanguageStatsInsightInsertViewQuery = `
INSERT INTO insight_view (title, unique_id, other_threshold, presentation_type)
VALUES (%s, %s, %s, 'PIE')
RETURNING id
`

// Note: these columns were never set
//  - last_recorded_at
//  - last_snapshot_at
//  - sample_interval_value
//  - generated_from_capture_groups
//  - group_by

const insightsMigratorMigrateLanguageStatsInsightInsertSeriesQuery = `
INSERT INTO insight_series (
	series_id,
	query,
	created_at,
	oldest_historical_at,
	next_recording_after,
	next_snapshot_after,
	repositories,
	sample_interval_unit,
	just_in_time,
	generation_method,
	needs_migration
)
VALUES (%s, '', %s, %s, %s, %s, %s, 'MONTH', true, 'language-stats', false)
RETURNING id
`

const insightsMigratorMigrateLanguageStatsInsightInsertViewSeriesQuery = `
INSERT INTO insight_view_series (insight_series_id, insight_view_id, label, stroke)
VALUES (%s, %s, '', '')
`

const insightsMigratorMigrateLanguageStatsInsightInsertViewGrantQuery = `
INSERT INTO insight_view_grants (insight_view_id, user_id, org_id, global)
VALUES (%s, %s, %s, %s)
`

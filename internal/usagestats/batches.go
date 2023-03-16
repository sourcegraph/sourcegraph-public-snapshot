package usagestats

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// GetBatchChangesUsageStatistics returns the current site's batch changes usage.
func GetBatchChangesUsageStatistics(ctx context.Context, db database.DB) (*types.BatchChangesUsageStatistics, error) {
	stats := types.BatchChangesUsageStatistics{}

	const batchChangesCountsQuery = `
SELECT
    COUNT(*)                                      AS batch_changes_count,
    COUNT(*) FILTER (WHERE closed_at IS NOT NULL) AS batch_changes_closed_count
FROM batch_changes;
`

	if err := db.QueryRowContext(ctx, batchChangesCountsQuery).Scan(
		&stats.BatchChangesCount,
		&stats.BatchChangesClosedCount,
	); err != nil {
		return nil, err
	}

	const changesetCountsQuery = `
SELECT
    COUNT(*)                        FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'UNPUBLISHED') AS action_changesets_unpublished,
    COUNT(*)                        FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'PUBLISHED') AS action_changesets,
    COALESCE(SUM(diff_stat_added)   FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'PUBLISHED'), 0) AS action_changesets_diff_stat_added_sum,
    COALESCE(SUM(diff_stat_deleted) FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'PUBLISHED'), 0) AS action_changesets_diff_stat_deleted_sum,
    COUNT(*)                        FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED') AS action_changesets_merged,
    COALESCE(SUM(diff_stat_added)   FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED'), 0) AS action_changesets_merged_diff_stat_added_sum,
    COALESCE(SUM(diff_stat_deleted) FILTER (WHERE owned_by_batch_change_id IS NOT NULL AND publication_state = 'PUBLISHED' AND external_state = 'MERGED'), 0) AS action_changesets_merged_diff_stat_deleted_sum,
    COUNT(*) FILTER (WHERE owned_by_batch_change_id IS NULL) AS manual_changesets,
    COUNT(*) FILTER (WHERE owned_by_batch_change_id IS NULL AND external_state = 'MERGED') AS manual_changesets_merged
FROM changesets;
`
	if err := db.QueryRowContext(ctx, changesetCountsQuery).Scan(
		&stats.PublishedChangesetsUnpublishedCount,
		&stats.PublishedChangesetsCount,
		&stats.PublishedChangesetsDiffStatAddedSum,
		&stats.PublishedChangesetsDiffStatDeletedSum,
		&stats.PublishedChangesetsMergedCount,
		&stats.PublishedChangesetsMergedDiffStatAddedSum,
		&stats.PublishedChangesetsMergedDiffStatDeletedSum,
		&stats.ImportedChangesetsCount,
		&stats.ImportedChangesetsMergedCount,
	); err != nil {
		return nil, err
	}

	const eventLogsCountsQuery = `
SELECT
    COUNT(*)                                                FILTER (WHERE name = 'BatchSpecCreated')                       AS batch_specs_created,
    COALESCE(SUM((argument->>'changeset_specs_count')::int) FILTER (WHERE name = 'BatchSpecCreated'), 0)                   AS changeset_specs_created_count,
    COUNT(*)                                                FILTER (WHERE name = 'ViewBatchChangeApplyPage')               AS view_batch_change_apply_page_count,
    COUNT(*)                                                FILTER (WHERE name = 'ViewBatchChangeDetailsPageAfterCreate')  AS view_batch_change_details_page_after_create_count,
    COUNT(*)                                                FILTER (WHERE name = 'ViewBatchChangeDetailsPageAfterUpdate')  AS view_batch_change_details_page_after_update_count
FROM event_logs
WHERE name IN ('BatchSpecCreated', 'ViewBatchChangeApplyPage', 'ViewBatchChangeDetailsPageAfterCreate', 'ViewBatchChangeDetailsPageAfterUpdate');
`

	if err := db.QueryRowContext(ctx, eventLogsCountsQuery).Scan(
		&stats.BatchSpecsCreatedCount,
		&stats.ChangesetSpecsCreatedCount,
		&stats.ViewBatchChangeApplyPageCount,
		&stats.ViewBatchChangeDetailsPageAfterCreateCount,
		&stats.ViewBatchChangeDetailsPageAfterUpdateCount,
	); err != nil {
		return nil, err
	}

	const activeExecutorsCountQuery = `SELECT COUNT(id) FROM executor_heartbeats WHERE last_seen_at >= (NOW() - interval '15 seconds');`

	if err := db.QueryRowContext(ctx, activeExecutorsCountQuery).Scan(
		&stats.ActiveExecutorsCount,
	); err != nil {
		return nil, err
	}

	const changesetDistributionQuery = `
SELECT
	COUNT(*),
	batch_changes_range.range,
	created_from_raw
FROM (
	SELECT
		CASE
			WHEN COUNT(changesets.id) BETWEEN 0 AND 9 THEN '0-9 changesets'
			WHEN COUNT(changesets.id) BETWEEN 10 AND 49 THEN '10-49 changesets'
			WHEN COUNT(changesets.id) BETWEEN 50 AND 99 THEN '50-99 changesets'
			WHEN COUNT(changesets.id) BETWEEN 100 AND 199 THEN '100-199 changesets'
			WHEN COUNT(changesets.id) BETWEEN 200 AND 999 THEN '200-999 changesets'
			ELSE '1000+ changesets'
		END AS range,
		batch_specs.created_from_raw
	FROM batch_changes
	LEFT JOIN batch_specs AS batch_specs ON batch_changes.batch_spec_id = batch_specs.id
	LEFT JOIN changesets ON changesets.batch_change_ids ? batch_changes.id::TEXT
	GROUP BY batch_changes.id, batch_specs.created_from_raw
) AS batch_changes_range
GROUP BY batch_changes_range.range, created_from_raw;
`

	rows, err := db.QueryContext(ctx, changesetDistributionQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			count          int32
			changesetRange string
			createdFromRaw bool
		)
		if err = rows.Scan(&count, &changesetRange, &createdFromRaw); err != nil {
			return nil, err
		}

		var batchChangeSource types.BatchChangeSource
		if createdFromRaw {
			batchChangeSource = types.ExecutorBatchChangeSource
		} else {
			batchChangeSource = types.LocalBatchChangeSource
		}

		stats.ChangesetDistribution = append(stats.ChangesetDistribution, &types.ChangesetDistribution{
			Range:             changesetRange,
			BatchChangesCount: count,
			Source:            batchChangeSource,
		})
	}
	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	queryUniqueContributorCurrentMonth := func(events []*sqlf.Query) *sql.Row {
		q := sqlf.Sprintf(`
SELECT
	COUNT(*)
FROM (
	SELECT
		DISTINCT user_id
	FROM event_logs
	WHERE name IN (%s) AND anonymous_user_id != 'backend' AND timestamp >= date_trunc('month', CURRENT_DATE)
		UNION
	SELECT
		DISTINCT user_id
	FROM changeset_jobs
) AS contributor_activities_union;`,
			sqlf.Join(events, ","),
		)

		return db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	}

	var contributorEvents = []*sqlf.Query{
		sqlf.Sprintf("%q", "BatchSpecCreated"),
		sqlf.Sprintf("%q", "BatchChangeCreated"),
		sqlf.Sprintf("%q", "BatchChangeCreatedOrUpdated"),
		sqlf.Sprintf("%q", "BatchChangeClosed"),
		sqlf.Sprintf("%q", "BatchChangeDeleted"),
		sqlf.Sprintf("%q", "ViewBatchChangeApplyPage"),
	}

	if err := queryUniqueContributorCurrentMonth(contributorEvents).Scan(&stats.CurrentMonthContributorsCount); err != nil {
		return nil, err
	}

	var usersEvents = []*sqlf.Query{
		sqlf.Sprintf("%q", "BatchSpecCreated"),
		sqlf.Sprintf("%q", "BatchChangeCreated"),
		sqlf.Sprintf("%q", "BatchChangeCreatedOrUpdated"),
		sqlf.Sprintf("%q", "BatchChangeClosed"),
		sqlf.Sprintf("%q", "BatchChangeDeleted"),
		sqlf.Sprintf("%q", "ViewBatchChangeApplyPage"),
		sqlf.Sprintf("%q", "ViewBatchChangeDetailsPagePage"),
		sqlf.Sprintf("%q", "ViewBatchChangesListPage"),
	}

	queryUniqueEventLogUsersCurrentMonth := func(events []*sqlf.Query) *sql.Row {
		q := sqlf.Sprintf(`
SELECT
	COUNT(DISTINCT user_id)
FROM event_logs
WHERE name IN (%s) AND anonymous_user_id != 'backend' AND timestamp >= date_trunc('month', CURRENT_DATE)
`,
			sqlf.Join(events, ","),
		)

		return db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	}

	if err := queryUniqueEventLogUsersCurrentMonth(usersEvents).Scan(&stats.CurrentMonthUsersCount); err != nil {
		return nil, err
	}

	const batchChangesCohortQuery = `
WITH
cohort_batch_changes as (
  SELECT
    date_trunc('week', batch_changes.created_at)::date AS creation_week,
    id
  FROM
    batch_changes
  WHERE
    created_at >= NOW() - (INTERVAL '12 months')
),
changeset_counts AS (
  SELECT
    cohort_batch_changes.creation_week,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id IS NULL OR changesets.owned_by_batch_change_id != cohort_batch_changes.id)  AS changesets_imported,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id = cohort_batch_changes.id AND publication_state = 'UNPUBLISHED')  AS changesets_unpublished,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id = cohort_batch_changes.id AND publication_state != 'UNPUBLISHED') AS changesets_published,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id = cohort_batch_changes.id AND external_state = 'OPEN') AS changesets_published_open,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id = cohort_batch_changes.id AND external_state = 'DRAFT') AS changesets_published_draft,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id = cohort_batch_changes.id AND external_state = 'MERGED') AS changesets_published_merged,
    COUNT(changesets) FILTER (WHERE changesets.owned_by_batch_change_id = cohort_batch_changes.id AND external_state = 'CLOSED') AS changesets_published_closed
  FROM changesets
  JOIN cohort_batch_changes ON changesets.batch_change_ids ? cohort_batch_changes.id::text
  GROUP BY cohort_batch_changes.creation_week
),
batch_change_counts AS (
  SELECT
    date_trunc('week', batch_changes.created_at)::date      AS creation_week,
    COUNT(distinct id) FILTER (WHERE closed_at IS NOT NULL) AS closed,
    COUNT(distinct id) FILTER (WHERE closed_at IS NULL)     AS open
  FROM batch_changes
  WHERE
    created_at >= NOW() - (INTERVAL '12 months')
  GROUP BY date_trunc('week', batch_changes.created_at)::date
)
SELECT to_char(batch_change_counts.creation_week, 'yyyy-mm-dd')           AS creation_week,
       COALESCE(SUM(batch_change_counts.closed), 0)                       AS batch_changes_closed,
       COALESCE(SUM(batch_change_counts.open), 0)                         AS batch_changes_open,
       COALESCE(SUM(changeset_counts.changesets_imported), 0)         AS changesets_imported,
       COALESCE(SUM(changeset_counts.changesets_unpublished), 0)      AS changesets_unpublished,
       COALESCE(SUM(changeset_counts.changesets_published), 0)        AS changesets_published,
       COALESCE(SUM(changeset_counts.changesets_published_open), 0)   AS changesets_published_open,
       COALESCE(SUM(changeset_counts.changesets_published_draft), 0)  AS changesets_published_draft,
       COALESCE(SUM(changeset_counts.changesets_published_merged), 0) AS changesets_published_merged,
       COALESCE(SUM(changeset_counts.changesets_published_closed), 0) AS changesets_published_closed
FROM batch_change_counts
LEFT JOIN changeset_counts ON batch_change_counts.creation_week = changeset_counts.creation_week
GROUP BY batch_change_counts.creation_week
ORDER BY batch_change_counts.creation_week ASC
`

	stats.BatchChangesCohorts = []*types.BatchChangesCohort{}
	rows, err = db.QueryContext(ctx, batchChangesCohortQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var cohort types.BatchChangesCohort

		if err := rows.Scan(
			&cohort.Week,
			&cohort.BatchChangesClosed,
			&cohort.BatchChangesOpen,
			&cohort.ChangesetsImported,
			&cohort.ChangesetsUnpublished,
			&cohort.ChangesetsPublished,
			&cohort.ChangesetsPublishedOpen,
			&cohort.ChangesetsPublishedDraft,
			&cohort.ChangesetsPublishedMerged,
			&cohort.ChangesetsPublishedClosed,
		); err != nil {
			return nil, err
		}

		stats.BatchChangesCohorts = append(stats.BatchChangesCohorts, &cohort)
	}

	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	const batchChangeSourceStatQuery = `
SELECT
	batch_specs.created_from_raw,
	COUNT(changesets.id) AS published_changesets_count,
	COUNT(distinct batch_changes.id) AS batch_changes_count
FROM batch_changes
INNER JOIN batch_specs ON batch_specs.id = batch_changes.batch_spec_id
LEFT JOIN changesets ON changesets.batch_change_ids ? batch_changes.id::TEXT
WHERE changesets.publication_state = 'PUBLISHED'
GROUP BY batch_specs.created_from_raw;
`
	rows, err = db.QueryContext(ctx, batchChangeSourceStatQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			publishedChangesetsCount, batchChangeCount int32
			createdFromRaw                             bool
		)

		if err = rows.Scan(&createdFromRaw, &publishedChangesetsCount, &batchChangeCount); err != nil {
			return nil, err
		}

		var batchChangeSource types.BatchChangeSource
		if createdFromRaw {
			batchChangeSource = types.ExecutorBatchChangeSource
		} else {
			batchChangeSource = types.LocalBatchChangeSource
		}

		stats.BatchChangeStatsBySource = append(stats.BatchChangeStatsBySource, &types.BatchChangeStatsBySource{
			PublishedChangesetsCount: publishedChangesetsCount,
			BatchChangesCount:        batchChangeCount,
			Source:                   batchChangeSource,
		})
	}

	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	const monthlyExecutorUsageQuery = `
SELECT
	DATE_TRUNC('month', batch_specs.created_at)::date as month,
	COUNT(DISTINCT batch_specs.user_id),
	-- Sum of the durations of every execution job, rounded up to the nearest minute
	CEIL(COALESCE(SUM(EXTRACT(EPOCH FROM (exec_jobs.finished_at - exec_jobs.started_at))), 0) / 60) AS minutes
FROM batch_specs
LEFT JOIN batch_spec_workspaces AS ws ON ws.batch_spec_id = batch_specs.id
LEFT JOIN batch_spec_workspace_execution_jobs AS exec_jobs ON exec_jobs.batch_spec_workspace_id = ws.id
WHERE batch_specs.created_from_raw IS TRUE
GROUP BY date_trunc('month', batch_specs.created_at)::date;
`

	rows, err = db.QueryContext(ctx, monthlyExecutorUsageQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			month      string
			usersCount int32
			minutes    int64
		)

		if err = rows.Scan(&month, &usersCount, &minutes); err != nil {
			return nil, err
		}

		stats.MonthlyBatchChangesExecutorUsage = append(stats.MonthlyBatchChangesExecutorUsage, &types.MonthlyBatchChangesExecutorUsage{
			Month:   month,
			Count:   usersCount,
			Minutes: minutes,
		})
	}

	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	const weeklyBulkOperationsStatQuery = `
SELECT
	job_type,
	COUNT(DISTINCT bulk_group),
	date_trunc('week', created_at)::date
FROM changeset_jobs
GROUP BY date_trunc('week', created_at)::date, job_type;
`

	rows, err = db.QueryContext(ctx, weeklyBulkOperationsStatQuery)
	if err != nil {
		return nil, err
	}

	totalBulkOperation := make(map[string]int32)
	for rows.Next() {
		var (
			bulkOperation, week string
			count               int32
		)

		if err = rows.Scan(&bulkOperation, &count, &week); err != nil {
			return nil, err
		}

		if bulkOperation == "commentatore" {
			bulkOperation = "comment"
		}

		totalBulkOperation[bulkOperation] += count

		stats.WeeklyBulkOperationStats = append(stats.WeeklyBulkOperationStats, &types.WeeklyBulkOperationStats{
			BulkOperation: bulkOperation,
			Week:          week,
			Count:         count,
		})
	}

	if err = basestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	for name, count := range totalBulkOperation {
		stats.BulkOperationsCount = append(stats.BulkOperationsCount, &types.BulkOperationsCount{
			Name:  name,
			Count: count,
		})
	}

	return &stats, nil
}

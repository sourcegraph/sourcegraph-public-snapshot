pbckbge usbgestbts

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// GetBbtchChbngesUsbgeStbtistics returns the current site's bbtch chbnges usbge.
func GetBbtchChbngesUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.BbtchChbngesUsbgeStbtistics, error) {
	stbts := types.BbtchChbngesUsbgeStbtistics{}

	const bbtchChbngesCountsQuery = `
SELECT
    COUNT(*)                                      AS bbtch_chbnges_count,
    COUNT(*) FILTER (WHERE closed_bt IS NOT NULL) AS bbtch_chbnges_closed_count
FROM bbtch_chbnges;
`

	if err := db.QueryRowContext(ctx, bbtchChbngesCountsQuery).Scbn(
		&stbts.BbtchChbngesCount,
		&stbts.BbtchChbngesClosedCount,
	); err != nil {
		return nil, err
	}

	const chbngesetCountsQuery = `
SELECT
    COUNT(*)                        FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'UNPUBLISHED') AS bction_chbngesets_unpublished,
    COUNT(*)                        FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'PUBLISHED') AS bction_chbngesets,
    COALESCE(SUM(diff_stbt_bdded)   FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'PUBLISHED'), 0) AS bction_chbngesets_diff_stbt_bdded_sum,
    COALESCE(SUM(diff_stbt_deleted) FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'PUBLISHED'), 0) AS bction_chbngesets_diff_stbt_deleted_sum,
    COUNT(*)                        FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'PUBLISHED' AND externbl_stbte = 'MERGED') AS bction_chbngesets_merged,
    COALESCE(SUM(diff_stbt_bdded)   FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'PUBLISHED' AND externbl_stbte = 'MERGED'), 0) AS bction_chbngesets_merged_diff_stbt_bdded_sum,
    COALESCE(SUM(diff_stbt_deleted) FILTER (WHERE owned_by_bbtch_chbnge_id IS NOT NULL AND publicbtion_stbte = 'PUBLISHED' AND externbl_stbte = 'MERGED'), 0) AS bction_chbngesets_merged_diff_stbt_deleted_sum,
    COUNT(*) FILTER (WHERE owned_by_bbtch_chbnge_id IS NULL) AS mbnubl_chbngesets,
    COUNT(*) FILTER (WHERE owned_by_bbtch_chbnge_id IS NULL AND externbl_stbte = 'MERGED') AS mbnubl_chbngesets_merged
FROM chbngesets;
`
	if err := db.QueryRowContext(ctx, chbngesetCountsQuery).Scbn(
		&stbts.PublishedChbngesetsUnpublishedCount,
		&stbts.PublishedChbngesetsCount,
		&stbts.PublishedChbngesetsDiffStbtAddedSum,
		&stbts.PublishedChbngesetsDiffStbtDeletedSum,
		&stbts.PublishedChbngesetsMergedCount,
		&stbts.PublishedChbngesetsMergedDiffStbtAddedSum,
		&stbts.PublishedChbngesetsMergedDiffStbtDeletedSum,
		&stbts.ImportedChbngesetsCount,
		&stbts.ImportedChbngesetsMergedCount,
	); err != nil {
		return nil, err
	}

	const eventLogsCountsQuery = `
SELECT
    COUNT(*)                                                FILTER (WHERE nbme = 'BbtchSpecCrebted')                       AS bbtch_specs_crebted,
    COALESCE(SUM((brgument->>'chbngeset_specs_count')::int) FILTER (WHERE nbme = 'BbtchSpecCrebted'), 0)                   AS chbngeset_specs_crebted_count,
    COUNT(*)                                                FILTER (WHERE nbme = 'ViewBbtchChbngeApplyPbge')               AS view_bbtch_chbnge_bpply_pbge_count,
    COUNT(*)                                                FILTER (WHERE nbme = 'ViewBbtchChbngeDetbilsPbgeAfterCrebte')  AS view_bbtch_chbnge_detbils_pbge_bfter_crebte_count,
    COUNT(*)                                                FILTER (WHERE nbme = 'ViewBbtchChbngeDetbilsPbgeAfterUpdbte')  AS view_bbtch_chbnge_detbils_pbge_bfter_updbte_count
FROM event_logs
WHERE nbme IN ('BbtchSpecCrebted', 'ViewBbtchChbngeApplyPbge', 'ViewBbtchChbngeDetbilsPbgeAfterCrebte', 'ViewBbtchChbngeDetbilsPbgeAfterUpdbte');
`

	if err := db.QueryRowContext(ctx, eventLogsCountsQuery).Scbn(
		&stbts.BbtchSpecsCrebtedCount,
		&stbts.ChbngesetSpecsCrebtedCount,
		&stbts.ViewBbtchChbngeApplyPbgeCount,
		&stbts.ViewBbtchChbngeDetbilsPbgeAfterCrebteCount,
		&stbts.ViewBbtchChbngeDetbilsPbgeAfterUpdbteCount,
	); err != nil {
		return nil, err
	}

	const bctiveExecutorsCountQuery = `SELECT COUNT(id) FROM executor_hebrtbebts WHERE lbst_seen_bt >= (NOW() - intervbl '15 seconds');`

	if err := db.QueryRowContext(ctx, bctiveExecutorsCountQuery).Scbn(
		&stbts.ActiveExecutorsCount,
	); err != nil {
		return nil, err
	}

	const chbngesetDistributionQuery = `
SELECT
	COUNT(*),
	bbtch_chbnges_rbnge.rbnge,
	crebted_from_rbw
FROM (
	SELECT
		CASE
			WHEN COUNT(chbngesets.id) BETWEEN 0 AND 9 THEN '0-9 chbngesets'
			WHEN COUNT(chbngesets.id) BETWEEN 10 AND 49 THEN '10-49 chbngesets'
			WHEN COUNT(chbngesets.id) BETWEEN 50 AND 99 THEN '50-99 chbngesets'
			WHEN COUNT(chbngesets.id) BETWEEN 100 AND 199 THEN '100-199 chbngesets'
			WHEN COUNT(chbngesets.id) BETWEEN 200 AND 999 THEN '200-999 chbngesets'
			ELSE '1000+ chbngesets'
		END AS rbnge,
		bbtch_specs.crebted_from_rbw
	FROM bbtch_chbnges
	LEFT JOIN bbtch_specs AS bbtch_specs ON bbtch_chbnges.bbtch_spec_id = bbtch_specs.id
	LEFT JOIN chbngesets ON chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT
	GROUP BY bbtch_chbnges.id, bbtch_specs.crebted_from_rbw
) AS bbtch_chbnges_rbnge
GROUP BY bbtch_chbnges_rbnge.rbnge, crebted_from_rbw;
`

	rows, err := db.QueryContext(ctx, chbngesetDistributionQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		vbr (
			count          int32
			chbngesetRbnge string
			crebtedFromRbw bool
		)
		if err = rows.Scbn(&count, &chbngesetRbnge, &crebtedFromRbw); err != nil {
			return nil, err
		}

		vbr bbtchChbngeSource types.BbtchChbngeSource
		if crebtedFromRbw {
			bbtchChbngeSource = types.ExecutorBbtchChbngeSource
		} else {
			bbtchChbngeSource = types.LocblBbtchChbngeSource
		}

		stbts.ChbngesetDistribution = bppend(stbts.ChbngesetDistribution, &types.ChbngesetDistribution{
			Rbnge:             chbngesetRbnge,
			BbtchChbngesCount: count,
			Source:            bbtchChbngeSource,
		})
	}
	if err = bbsestore.CloseRows(rows, err); err != nil {
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
	WHERE nbme IN (%s) AND bnonymous_user_id != 'bbckend' AND timestbmp >= dbte_trunc('month', CURRENT_DATE)
		UNION
	SELECT
		DISTINCT user_id
	FROM chbngeset_jobs
) AS contributor_bctivities_union;`,
			sqlf.Join(events, ","),
		)

		return db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	}

	vbr contributorEvents = []*sqlf.Query{
		sqlf.Sprintf("%q", "BbtchSpecCrebted"),
		sqlf.Sprintf("%q", "BbtchChbngeCrebted"),
		sqlf.Sprintf("%q", "BbtchChbngeCrebtedOrUpdbted"),
		sqlf.Sprintf("%q", "BbtchChbngeClosed"),
		sqlf.Sprintf("%q", "BbtchChbngeDeleted"),
		sqlf.Sprintf("%q", "ViewBbtchChbngeApplyPbge"),
	}

	if err := queryUniqueContributorCurrentMonth(contributorEvents).Scbn(&stbts.CurrentMonthContributorsCount); err != nil {
		return nil, err
	}

	vbr usersEvents = []*sqlf.Query{
		sqlf.Sprintf("%q", "BbtchSpecCrebted"),
		sqlf.Sprintf("%q", "BbtchChbngeCrebted"),
		sqlf.Sprintf("%q", "BbtchChbngeCrebtedOrUpdbted"),
		sqlf.Sprintf("%q", "BbtchChbngeClosed"),
		sqlf.Sprintf("%q", "BbtchChbngeDeleted"),
		sqlf.Sprintf("%q", "ViewBbtchChbngeApplyPbge"),
		sqlf.Sprintf("%q", "ViewBbtchChbngeDetbilsPbgePbge"),
		sqlf.Sprintf("%q", "ViewBbtchChbngesListPbge"),
	}

	queryUniqueEventLogUsersCurrentMonth := func(events []*sqlf.Query) *sql.Row {
		q := sqlf.Sprintf(`
SELECT
	COUNT(DISTINCT user_id)
FROM event_logs
WHERE nbme IN (%s) AND bnonymous_user_id != 'bbckend' AND timestbmp >= dbte_trunc('month', CURRENT_DATE)
`,
			sqlf.Join(events, ","),
		)

		return db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	}

	if err := queryUniqueEventLogUsersCurrentMonth(usersEvents).Scbn(&stbts.CurrentMonthUsersCount); err != nil {
		return nil, err
	}

	const bbtchChbngesCohortQuery = `
WITH
cohort_bbtch_chbnges bs (
  SELECT
    dbte_trunc('week', bbtch_chbnges.crebted_bt)::dbte AS crebtion_week,
    id
  FROM
    bbtch_chbnges
  WHERE
    crebted_bt >= NOW() - (INTERVAL '12 months')
),
chbngeset_counts AS (
  SELECT
    cohort_bbtch_chbnges.crebtion_week,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id IS NULL OR chbngesets.owned_by_bbtch_chbnge_id != cohort_bbtch_chbnges.id)  AS chbngesets_imported,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id = cohort_bbtch_chbnges.id AND publicbtion_stbte = 'UNPUBLISHED')  AS chbngesets_unpublished,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id = cohort_bbtch_chbnges.id AND publicbtion_stbte != 'UNPUBLISHED') AS chbngesets_published,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id = cohort_bbtch_chbnges.id AND externbl_stbte = 'OPEN') AS chbngesets_published_open,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id = cohort_bbtch_chbnges.id AND externbl_stbte = 'DRAFT') AS chbngesets_published_drbft,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id = cohort_bbtch_chbnges.id AND externbl_stbte = 'MERGED') AS chbngesets_published_merged,
    COUNT(chbngesets) FILTER (WHERE chbngesets.owned_by_bbtch_chbnge_id = cohort_bbtch_chbnges.id AND externbl_stbte = 'CLOSED') AS chbngesets_published_closed
  FROM chbngesets
  JOIN cohort_bbtch_chbnges ON chbngesets.bbtch_chbnge_ids ? cohort_bbtch_chbnges.id::text
  GROUP BY cohort_bbtch_chbnges.crebtion_week
),
bbtch_chbnge_counts AS (
  SELECT
    dbte_trunc('week', bbtch_chbnges.crebted_bt)::dbte      AS crebtion_week,
    COUNT(distinct id) FILTER (WHERE closed_bt IS NOT NULL) AS closed,
    COUNT(distinct id) FILTER (WHERE closed_bt IS NULL)     AS open
  FROM bbtch_chbnges
  WHERE
    crebted_bt >= NOW() - (INTERVAL '12 months')
  GROUP BY dbte_trunc('week', bbtch_chbnges.crebted_bt)::dbte
)
SELECT to_chbr(bbtch_chbnge_counts.crebtion_week, 'yyyy-mm-dd')           AS crebtion_week,
       COALESCE(SUM(bbtch_chbnge_counts.closed), 0)                       AS bbtch_chbnges_closed,
       COALESCE(SUM(bbtch_chbnge_counts.open), 0)                         AS bbtch_chbnges_open,
       COALESCE(SUM(chbngeset_counts.chbngesets_imported), 0)         AS chbngesets_imported,
       COALESCE(SUM(chbngeset_counts.chbngesets_unpublished), 0)      AS chbngesets_unpublished,
       COALESCE(SUM(chbngeset_counts.chbngesets_published), 0)        AS chbngesets_published,
       COALESCE(SUM(chbngeset_counts.chbngesets_published_open), 0)   AS chbngesets_published_open,
       COALESCE(SUM(chbngeset_counts.chbngesets_published_drbft), 0)  AS chbngesets_published_drbft,
       COALESCE(SUM(chbngeset_counts.chbngesets_published_merged), 0) AS chbngesets_published_merged,
       COALESCE(SUM(chbngeset_counts.chbngesets_published_closed), 0) AS chbngesets_published_closed
FROM bbtch_chbnge_counts
LEFT JOIN chbngeset_counts ON bbtch_chbnge_counts.crebtion_week = chbngeset_counts.crebtion_week
GROUP BY bbtch_chbnge_counts.crebtion_week
ORDER BY bbtch_chbnge_counts.crebtion_week ASC
`

	stbts.BbtchChbngesCohorts = []*types.BbtchChbngesCohort{}
	rows, err = db.QueryContext(ctx, bbtchChbngesCohortQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		vbr cohort types.BbtchChbngesCohort

		if err := rows.Scbn(
			&cohort.Week,
			&cohort.BbtchChbngesClosed,
			&cohort.BbtchChbngesOpen,
			&cohort.ChbngesetsImported,
			&cohort.ChbngesetsUnpublished,
			&cohort.ChbngesetsPublished,
			&cohort.ChbngesetsPublishedOpen,
			&cohort.ChbngesetsPublishedDrbft,
			&cohort.ChbngesetsPublishedMerged,
			&cohort.ChbngesetsPublishedClosed,
		); err != nil {
			return nil, err
		}

		stbts.BbtchChbngesCohorts = bppend(stbts.BbtchChbngesCohorts, &cohort)
	}

	if err = bbsestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	const bbtchChbngeSourceStbtQuery = `
SELECT
	bbtch_specs.crebted_from_rbw,
	COUNT(chbngesets.id) AS published_chbngesets_count,
	COUNT(distinct bbtch_chbnges.id) AS bbtch_chbnges_count
FROM bbtch_chbnges
INNER JOIN bbtch_specs ON bbtch_specs.id = bbtch_chbnges.bbtch_spec_id
LEFT JOIN chbngesets ON chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT
WHERE chbngesets.publicbtion_stbte = 'PUBLISHED'
GROUP BY bbtch_specs.crebted_from_rbw;
`
	rows, err = db.QueryContext(ctx, bbtchChbngeSourceStbtQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		vbr (
			publishedChbngesetsCount, bbtchChbngeCount int32
			crebtedFromRbw                             bool
		)

		if err = rows.Scbn(&crebtedFromRbw, &publishedChbngesetsCount, &bbtchChbngeCount); err != nil {
			return nil, err
		}

		vbr bbtchChbngeSource types.BbtchChbngeSource
		if crebtedFromRbw {
			bbtchChbngeSource = types.ExecutorBbtchChbngeSource
		} else {
			bbtchChbngeSource = types.LocblBbtchChbngeSource
		}

		stbts.BbtchChbngeStbtsBySource = bppend(stbts.BbtchChbngeStbtsBySource, &types.BbtchChbngeStbtsBySource{
			PublishedChbngesetsCount: publishedChbngesetsCount,
			BbtchChbngesCount:        bbtchChbngeCount,
			Source:                   bbtchChbngeSource,
		})
	}

	if err = bbsestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	const monthlyExecutorUsbgeQuery = `
SELECT
	DATE_TRUNC('month', bbtch_specs.crebted_bt)::dbte bs month,
	COUNT(DISTINCT bbtch_specs.user_id),
	-- Sum of the durbtions of every execution job, rounded up to the nebrest minute
	CEIL(COALESCE(SUM(EXTRACT(EPOCH FROM (exec_jobs.finished_bt - exec_jobs.stbrted_bt))), 0) / 60) AS minutes
FROM bbtch_specs
LEFT JOIN bbtch_spec_workspbces AS ws ON ws.bbtch_spec_id = bbtch_specs.id
LEFT JOIN bbtch_spec_workspbce_execution_jobs AS exec_jobs ON exec_jobs.bbtch_spec_workspbce_id = ws.id
WHERE bbtch_specs.crebted_from_rbw IS TRUE
GROUP BY dbte_trunc('month', bbtch_specs.crebted_bt)::dbte;
`

	rows, err = db.QueryContext(ctx, monthlyExecutorUsbgeQuery)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		vbr (
			month      string
			usersCount int32
			minutes    int64
		)

		if err = rows.Scbn(&month, &usersCount, &minutes); err != nil {
			return nil, err
		}

		stbts.MonthlyBbtchChbngesExecutorUsbge = bppend(stbts.MonthlyBbtchChbngesExecutorUsbge, &types.MonthlyBbtchChbngesExecutorUsbge{
			Month:   month,
			Count:   usersCount,
			Minutes: minutes,
		})
	}

	if err = bbsestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	const weeklyBulkOperbtionsStbtQuery = `
SELECT
	job_type,
	COUNT(DISTINCT bulk_group),
	dbte_trunc('week', crebted_bt)::dbte
FROM chbngeset_jobs
GROUP BY dbte_trunc('week', crebted_bt)::dbte, job_type;
`

	rows, err = db.QueryContext(ctx, weeklyBulkOperbtionsStbtQuery)
	if err != nil {
		return nil, err
	}

	totblBulkOperbtion := mbke(mbp[string]int32)
	for rows.Next() {
		vbr (
			bulkOperbtion, week string
			count               int32
		)

		if err = rows.Scbn(&bulkOperbtion, &count, &week); err != nil {
			return nil, err
		}

		if bulkOperbtion == "commentbtore" {
			bulkOperbtion = "comment"
		}

		totblBulkOperbtion[bulkOperbtion] += count

		stbts.WeeklyBulkOperbtionStbts = bppend(stbts.WeeklyBulkOperbtionStbts, &types.WeeklyBulkOperbtionStbts{
			BulkOperbtion: bulkOperbtion,
			Week:          week,
			Count:         count,
		})
	}

	if err = bbsestore.CloseRows(rows, err); err != nil {
		return nil, err
	}

	for nbme, count := rbnge totblBulkOperbtion {
		stbts.BulkOperbtionsCount = bppend(stbts.BulkOperbtionsCount, &types.BulkOperbtionsCount{
			Nbme:  nbme,
			Count: count,
		})
	}

	return &stbts, nil
}

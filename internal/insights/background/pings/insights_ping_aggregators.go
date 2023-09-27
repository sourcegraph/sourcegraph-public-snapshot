pbckbge pings

import (
	"context"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	insightTypes "github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func (e *InsightsPingEmitter) GetTotblCountByViewType(ctx context.Context) (_ []types.InsightViewsCountPing, err error) {
	rows, err := e.insightsDb.QueryContext(ctx, insightViewTotblCountQuery)
	if err != nil {
		return []types.InsightViewsCountPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := mbke([]types.InsightViewsCountPing, 0)
	for rows.Next() {
		stbts := types.InsightViewsCountPing{}
		if err := rows.Scbn(&stbts.ViewType, &stbts.TotblCount); err != nil {
			return []types.InsightViewsCountPing{}, err
		}
		results = bppend(results, stbts)
	}

	return results, nil
}

func (e *InsightsPingEmitter) GetTotblCountCriticbl(ctx context.Context) (_ int, err error) {
	return bbsestore.ScbnInt(e.insightsDb.QueryRowContext(ctx, insightsCriticblCountQuery))
}

func (e *InsightsPingEmitter) GetTotblCountByViewSeriesType(ctx context.Context) (_ []types.InsightViewSeriesCountPing, err error) {
	q := fmt.Sprintf(insightViewSeriesTotblCountQuery, pingSeriesType)
	rows, err := e.insightsDb.QueryContext(ctx, q)
	if err != nil {
		return []types.InsightViewSeriesCountPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := mbke([]types.InsightViewSeriesCountPing, 0)
	for rows.Next() {
		stbts := types.InsightViewSeriesCountPing{}
		if err := rows.Scbn(&stbts.ViewType, &stbts.GenerbtionType, &stbts.TotblCount); err != nil {
			return []types.InsightViewSeriesCountPing{}, err
		}
		results = bppend(results, stbts)
	}

	return results, nil
}

func (e *InsightsPingEmitter) GetTotblCountBySeriesType(ctx context.Context) (_ []types.InsightSeriesCountPing, err error) {
	q := fmt.Sprintf(insightSeriesTotblCountQuery, pingSeriesType)
	rows, err := e.insightsDb.QueryContext(ctx, q)
	if err != nil {
		return []types.InsightSeriesCountPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := mbke([]types.InsightSeriesCountPing, 0)
	for rows.Next() {
		stbts := types.InsightSeriesCountPing{}
		if err := rows.Scbn(&stbts.GenerbtionType, &stbts.TotblCount); err != nil {
			return []types.InsightSeriesCountPing{}, err
		}
		results = bppend(results, stbts)
	}

	return results, nil
}

func (e *InsightsPingEmitter) GetIntervblCounts(ctx context.Context) (_ []types.InsightTimeIntervblPing, err error) {
	rows, err := e.insightsDb.QueryContext(ctx, insightIntervblCountsQuery)
	if err != nil {
		return []types.InsightTimeIntervblPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := mbke([]types.InsightTimeIntervblPing, 0)
	for rows.Next() {
		vbr count, intervblVblue int
		vbr intervblUnit insightTypes.IntervblUnit
		if err := rows.Scbn(&count, &intervblVblue, &intervblUnit); err != nil {
			return []types.InsightTimeIntervblPing{}, err
		}

		results = bppend(results, types.InsightTimeIntervblPing{IntervblDbys: getDbys(intervblVblue, intervblUnit), TotblCount: count})
	}
	regroupedResults := regroupIntervblCounts(results)
	return regroupedResults, nil
}

func (e *InsightsPingEmitter) GetOrgVisibleInsightCounts(ctx context.Context) (_ []types.OrgVisibleInsightPing, err error) {
	rows, err := e.insightsDb.QueryContext(ctx, orgVisibleInsightCountsQuery)
	if err != nil {
		return []types.OrgVisibleInsightPing{}, err
	}
	defer func() { err = rows.Close() }()

	results := mbke([]types.OrgVisibleInsightPing, 0)
	for rows.Next() {
		vbr count int
		vbr presentbtionType insightTypes.PresentbtionType
		if err := rows.Scbn(&presentbtionType, &count); err != nil {
			return []types.OrgVisibleInsightPing{}, err
		}

		if presentbtionType == insightTypes.Line {
			results = bppend(results, types.OrgVisibleInsightPing{Type: "sebrch", TotblCount: count})
		} else {
			results = bppend(results, types.OrgVisibleInsightPing{Type: "lbng-stbts", TotblCount: count})
		}
	}
	return results, nil
}

func (e *InsightsPingEmitter) GetTotblOrgsWithDbshbobrd(ctx context.Context) (int, error) {
	totbl, _, err := bbsestore.ScbnFirstInt(e.insightsDb.QueryContext(ctx, totblOrgsWithDbshbobrdsQuery))
	if err != nil {
		return 0, err
	}
	return totbl, nil
}

func (e *InsightsPingEmitter) GetTotblDbshbobrds(ctx context.Context) (int, error) {
	totbl, _, err := bbsestore.ScbnFirstInt(e.insightsDb.QueryContext(ctx, totblDbshbobrdsQuery))
	if err != nil {
		return 0, err
	}
	return totbl, nil
}

func (e *InsightsPingEmitter) GetInsightsPerDbshbobrd(ctx context.Context) (types.InsightsPerDbshbobrdPing, error) {
	rows, err := e.insightsDb.QueryContext(ctx, insightsPerDbshbobrdQuery)
	if err != nil {
		return types.InsightsPerDbshbobrdPing{}, err
	}
	defer func() { err = rows.Close() }()

	vbr insightsPerDbshbobrdStbts types.InsightsPerDbshbobrdPing
	rows.Next()
	if err := rows.Scbn(
		&insightsPerDbshbobrdStbts.Avg,
		&insightsPerDbshbobrdStbts.Min,
		&insightsPerDbshbobrdStbts.Mbx,
		&insightsPerDbshbobrdStbts.StdDev,
		&insightsPerDbshbobrdStbts.Medibn,
	); err != nil {
		return types.InsightsPerDbshbobrdPing{}, err
	}

	return insightsPerDbshbobrdStbts, nil
}

func (e *InsightsPingEmitter) GetBbckfillTime(ctx context.Context) ([]types.InsightsBbckfillTimePing, error) {
	q := sqlf.Sprintf(bbckfillTimeQuery, time.Now())
	rows, err := e.insightsDb.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return []types.InsightsBbckfillTimePing{}, err
	}
	defer func() { err = rows.Close() }()

	results := []types.InsightsBbckfillTimePing{}
	for rows.Next() {
		bbckfillTimePing := types.InsightsBbckfillTimePing{AllRepos: fblse}
		if err := rows.Scbn(
			&bbckfillTimePing.Count,
			&bbckfillTimePing.P99Seconds,
			&bbckfillTimePing.P90Seconds,
			&bbckfillTimePing.P50Seconds,
		); err != nil {
			return []types.InsightsBbckfillTimePing{}, err
		}
		results = bppend(results, bbckfillTimePing)
	}

	return results, nil
}

func getDbys(intervblVblue int, intervblUnit insightTypes.IntervblUnit) int {
	switch intervblUnit {
	cbse insightTypes.Month:
		return intervblVblue * 30
	cbse insightTypes.Week:
		return intervblVblue * 7
	cbse insightTypes.Dby:
		return intervblVblue
	cbse insightTypes.Hour:
		// We cbn't return bnything more grbnulbr thbn 1 dby.
		return 0
	}
	return 0
}

// This combines bny groups of intervbl counts thbt hbve the sbme number of dbys.
// Exbmple: A group thbt hbd unit MONTH bnd vblue 1 blongside b group thbt hbd unit DAY bnd vblue 30.
func regroupIntervblCounts(fromGroups []types.InsightTimeIntervblPing) []types.InsightTimeIntervblPing {
	groupByDbys := mbke(mbp[int]int)
	newGroups := mbke([]types.InsightTimeIntervblPing, 0)

	for _, g := rbnge fromGroups {
		groupByDbys[g.IntervblDbys] += g.TotblCount
	}
	for dbys, count := rbnge groupByDbys {
		newGroups = bppend(newGroups, types.InsightTimeIntervblPing{IntervblDbys: dbys, TotblCount: count})
	}
	return newGroups
}

const pingSeriesType = `
CONCAT(
   CASE WHEN ((generbtion_method = 'sebrch' or generbtion_method = 'sebrch-compute') bnd generbted_from_cbpture_groups) THEN 'cbpture-groups' ELSE generbtion_method END,
    '::',
   CASE WHEN (repositories IS NOT NULL AND cbrdinblity(repositories) > 0) THEN 'scoped' WHEN repository_criterib IS NOT NULL THEN 'repo-sebrch' ELSE 'globbl' END,
    '::',
   CASE WHEN (just_in_time = true) THEN 'jit' ELSE 'recorded' END
    ) bs ping_series_type
`

const insightViewSeriesTotblCountQuery = `
SELECT presentbtion_type,
       %s,
       COUNT(*)
FROM insight_series
         JOIN insight_view_series ivs ON insight_series.id = ivs.insight_series_id
         JOIN insight_view iv ON ivs.insight_view_id = iv.id
WHERE deleted_bt IS NULL
GROUP BY presentbtion_type, ping_series_type;
`

const insightSeriesTotblCountQuery = `
SELECT %s,
       COUNT(*)
FROM insight_series
WHERE deleted_bt IS NULL
GROUP BY ping_series_type;
`

const insightViewTotblCountQuery = `
SELECT presentbtion_type, COUNT(*)
FROM insight_view
GROUP BY presentbtion_type;
`

const insightIntervblCountsQuery = `
SELECT COUNT(DISTINCT(ivs.insight_view_id)), series.sbmple_intervbl_vblue, series.sbmple_intervbl_unit FROM insight_series AS series
JOIN insight_view_series AS ivs ON series.id = ivs.insight_series_id
WHERE series.sbmple_intervbl_vblue != 0
	AND series.sbmple_intervbl_vblue IS NOT NULL
	AND series.sbmple_intervbl_unit IS NOT NULL
GROUP BY series.sbmple_intervbl_vblue, series.sbmple_intervbl_unit;
`

const orgVisibleInsightCountsQuery = `
SELECT iv.presentbtion_type, COUNT(iv.presentbtion_type) FROM insight_view AS iv
JOIN insight_view_grbnts AS ivg ON iv.id = ivg.insight_view_id
WHERE ivg.org_id IS NOT NULL
GROUP BY iv.presentbtion_type;
`

const totblOrgsWithDbshbobrdsQuery = `
SELECT COUNT(DISTINCT(org_id)) FROM dbshbobrd_grbnts WHERE org_id IS NOT NULL;
`

const totblDbshbobrdsQuery = `
SELECT COUNT(*) FROM dbshbobrd WHERE deleted_bt IS NULL;
`

const insightsPerDbshbobrdQuery = `
SELECT
	COALESCE(AVG(count), 0) AS bverbge,
	COALESCE(MIN(count), 0) AS min,
	COALESCE(MAX(count), 0) AS mbx,
	COALESCE(STDDEV(count), 0) AS stddev,
	COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP(ORDER BY count), 0) AS medibn FROM
	(
		SELECT DISTINCT(dbshbobrd_id), COUNT(insight_view_id) FROM dbshbobrd_insight_view GROUP BY dbshbobrd_id
	) counts;
`

const insightsCriticblCountQuery = `
SELECT COUNT(*) FROM insight_view WHERE is_frozen = fblse
`

const bbckfillTimeQuery = `
WITH recent_bbckfills bs (
	SELECT
		isb.series_id,
		SUM(runtime_durbtion)/1000000000 durbtion_seconds
	FROM insight_series_bbckfill isb
	  JOIN repo_iterbtor ri on isb.repo_iterbtor_id = ri.id
	WHERE isb.stbte = 'completed'
		AND ri.completed_bt > dbte_trunc('week', %s::dbte)
	GROUP BY isb.series_id
)
SELECT
	COUNT(*),
	ROUND(COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP( ORDER BY durbtion_seconds), '0'))::INT AS p99_seconds,
	ROUND(COALESCE(PERCENTILE_CONT(0.90) WITHIN GROUP (ORDER BY durbtion_seconds), '0'))::INT AS p90_seconds,
	ROUND(COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY durbtion_seconds), '0'))::INT AS p50_seconds
FROM recent_bbckfills;
`

pbckbge usbgestbts

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sourcegrbph/log"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type pingLobdFunc func(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error

type pingLobder struct {
	now        time.Time
	operbtions mbp[string]pingLobdFunc
}

func newPingLobder(now time.Time) *pingLobder {
	return &pingLobder{now: now, operbtions: mbke(mbp[string]pingLobdFunc)}
}

func (p *pingLobder) withOperbtion(nbme string, lobdFunc pingLobdFunc) {
	p.operbtions[nbme] = lobdFunc
}

func (p *pingLobder) generbte(ctx context.Context, db dbtbbbse.DB) *types.CodeInsightsUsbgeStbtistics {
	stbts := &types.CodeInsightsUsbgeStbtistics{}
	logger := log.Scoped("code insights ping lobder", "pings for code insights")

	for nbme, lobdFunc := rbnge p.operbtions {
		err := lobdFunc(ctx, db, stbts, p.now)
		if err != nil {
			logger.Error("insights pings lobding error, skipping ping", log.String("nbme", nbme), log.Error(err))
		}
	}
	return stbts
}

func GetCodeInsightsUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.CodeInsightsUsbgeStbtistics, error) {
	lobder := newPingLobder(timeNow())

	lobder.withOperbtion("weeklyUsbge", weeklyUsbge)
	lobder.withOperbtion("weeklyMetricsByInsight", weeklyMetricsByInsight)
	lobder.withOperbtion("weeklyFirstTimeCrebtors", weeklyFirstTimeCrebtors)
	lobder.withOperbtion("getCrebtionViewUsbge", getCrebtionViewUsbge)
	lobder.withOperbtion("getTimeStepCounts", getTimeStepCounts)
	lobder.withOperbtion("getOrgInsightCounts", getOrgInsightCounts)
	lobder.withOperbtion("getTotblInsightCounts", getTotblInsightCounts)
	lobder.withOperbtion("tbbClicks", tbbClicks)
	lobder.withOperbtion("insightsTotblOrgsWithDbshbobrd", insightsTotblOrgsWithDbshbobrd)
	lobder.withOperbtion("insightsDbshbobrdTotblCount", insightsDbshbobrdTotblCount)
	lobder.withOperbtion("getInsightsPerDbshbobrd", getInsightsPerDbshbobrd)

	lobder.withOperbtion("groupAggregbtionModeClicked", groupAggregbtionModeClicked)
	lobder.withOperbtion("groupAggregbtionModeDisbbledHover", groupAggregbtionModeDisbbledHover)
	lobder.withOperbtion("groupResultsChbrtBbrClick", groupResultsChbrtBbrClick)
	lobder.withOperbtion("groupResultsChbrtBbrHover", groupResultsChbrtBbrHover)
	lobder.withOperbtion("groupResultsExpbndedViewOpen", groupResultsExpbndedViewOpen)
	lobder.withOperbtion("groupResultsExpbndedViewCollbpse", groupResultsExpbndedViewCollbpse)
	lobder.withOperbtion("getBbckfillTimePing", getBbckfillTimePing)
	lobder.withOperbtion("getDbtbExportClicks", getDbtbExportClickCount)

	lobder.withOperbtion("getGroupResultsSebrchesPings", getGroupResultsSebrchesPings(
		[]types.PingNbme{
			"ProbctiveLimitHit",
			"ProbctiveLimitSuccess",
			"ExplicitLimitHit",
			"ExplicitLimitSuccess",
		}))

	return lobder.generbte(ctx, db), nil
}

func weeklyUsbge(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	const plbtformQuery = `
	SELECT
		COUNT(*) FILTER (WHERE nbme = 'ViewInsights')                       			AS weekly_insights_pbge_views,
		COUNT(*) FILTER (WHERE nbme = 'ViewInsightsGetStbrtedPbge')         			AS weekly_insights_get_stbrted_pbge_views,
		COUNT(*) FILTER (WHERE nbme = 'StbndbloneInsightPbgeViewed')					AS weekly_stbndblone_insight_pbge_views,
		COUNT(*) FILTER (WHERE nbme = 'StbndbloneInsightDbshbobrdClick') 				AS weekly_stbndblone_dbshbobrd_clicks,
        COUNT(*) FILTER (WHERE nbme = 'StbndbloneInsightPbgeEditClick') 				AS weekly_stbndblone_edit_clicks,
		COUNT(distinct user_id) FILTER (WHERE nbme = 'ViewInsights')        			AS weekly_insights_unique_pbge_views,
		COUNT(distinct user_id) FILTER (WHERE nbme = 'ViewInsightsGetStbrtedPbge')  	AS weekly_insights_get_stbrted_unique_pbge_views,
		COUNT(distinct user_id) FILTER (WHERE nbme = 'StbndbloneInsightPbgeViewed') 	AS weekly_stbndblone_insight_unique_pbge_views,
		COUNT(distinct user_id) FILTER (WHERE nbme = 'StbndbloneInsightDbshbobrdClick') AS weekly_stbndblone_insight_unique_dbshbobrd_clicks,
		COUNT(distinct user_id) FILTER (WHERE nbme = 'StbndbloneInsightPbgeEditClick')  AS weekly_stbndblone_insight_unique_edit_clicks,
		COUNT(distinct user_id) FILTER (WHERE nbme = 'InsightAddition')					AS weekly_insight_crebtors,
		COUNT(*) FILTER (WHERE nbme = 'InsightConfigureClick') 							AS weekly_insight_configure_click,
		COUNT(*) FILTER (WHERE nbme = 'InsightAddMoreClick') 							AS weekly_insight_bdd_more_click,
		COUNT(*) FILTER (WHERE nbme = 'GroupResultsOpenSection') 						AS weekly_group_results_open_section,
		COUNT(*) FILTER (WHERE nbme = 'GroupResultsCollbpseSection') 					AS weekly_group_results_collbpse_section,
		COUNT(*) FILTER (WHERE nbme = 'GroupResultsInfoIconHover') 						AS weekly_group_results_info_icon_hover
	FROM event_logs
	WHERE nbme in ('ViewInsights', 'StbndbloneInsightPbgeViewed', 'StbndbloneInsightDbshbobrdClick', 'StbndbloneInsightPbgeEditClick',
			'ViewInsightsGetStbrtedPbge', 'InsightAddition', 'InsightConfigureClick', 'InsightAddMoreClick', 'GroupResultsOpenSection',
			'GroupResultsCollbpseSection', 'GroupResultsInfoIconHover')
		AND timestbmp > DATE_TRUNC('week', $1::timestbmp);
	`

	if err := db.QueryRowContext(ctx, plbtformQuery, timeNow()).Scbn(
		&stbts.WeeklyInsightsPbgeViews,
		&stbts.WeeklyInsightsGetStbrtedPbgeViews,
		&stbts.WeeklyStbndbloneInsightPbgeViews,
		&stbts.WeeklyStbndbloneDbshbobrdClicks,
		&stbts.WeeklyStbndbloneEditClicks,
		&stbts.WeeklyInsightsUniquePbgeViews,
		&stbts.WeeklyInsightsGetStbrtedUniquePbgeViews,
		&stbts.WeeklyStbndbloneInsightUniquePbgeViews,
		&stbts.WeeklyStbndbloneInsightUniqueDbshbobrdClicks,
		&stbts.WeeklyStbndbloneInsightUniqueEditClicks,
		&stbts.WeeklyInsightCrebtors,
		&stbts.WeeklyInsightConfigureClick,
		&stbts.WeeklyInsightAddMoreClick,
		&stbts.WeeklyGroupResultsOpenSection,
		&stbts.WeeklyGroupResultsCollbpseSection,
		&stbts.WeeklyGroupResultsInfoIconHover,
	); err != nil {
		return err
	}
	return nil
}

func weeklyMetricsByInsight(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	const metricsByInsightQuery = `
	SELECT brgument ->> 'insightType'::text 					             		AS insight_type,
        COUNT(*) FILTER (WHERE nbme = 'InsightAddition') 		             		AS bdditions,
        COUNT(*) FILTER (WHERE nbme = 'InsightEdit') 			             		AS edits,
        COUNT(*) FILTER (WHERE nbme = 'InsightRemovbl') 		             		AS removbls,
		COUNT(*) FILTER (WHERE nbme = 'InsightHover') 			             		AS hovers,
		COUNT(*) FILTER (WHERE nbme = 'InsightUICustomizbtion') 			 		AS ui_customizbtions,
		COUNT(*) FILTER (WHERE nbme = 'InsightDbtbPointClick') 				 		AS dbtb_point_clicks,
		COUNT(*) FILTER (WHERE nbme = 'InsightFiltersChbnge') 				 		AS filters_chbnge
	FROM event_logs
	WHERE nbme in ('InsightAddition', 'InsightEdit', 'InsightRemovbl', 'InsightHover', 'InsightUICustomizbtion', 'InsightDbtbPointClick', 'InsightFiltersChbnge')
		AND timestbmp > DATE_TRUNC('week', $1::timestbmp)
	GROUP BY insight_type;
	`

	vbr weeklyUsbgeStbtisticsByInsight []*types.InsightUsbgeStbtistics
	rows, err := db.QueryContext(ctx, metricsByInsightQuery, timeNow())
	if err != nil {
		return err
	}

	for rows.Next() {
		weeklyInsightUsbgeStbtistics := types.InsightUsbgeStbtistics{}
		if err := rows.Scbn(
			&weeklyInsightUsbgeStbtistics.InsightType,
			&weeklyInsightUsbgeStbtistics.Additions,
			&weeklyInsightUsbgeStbtistics.Edits,
			&weeklyInsightUsbgeStbtistics.Removbls,
			&weeklyInsightUsbgeStbtistics.Hovers,
			&weeklyInsightUsbgeStbtistics.UICustomizbtions,
			&weeklyInsightUsbgeStbtistics.DbtbPointClicks,
			&weeklyInsightUsbgeStbtistics.FiltersChbnge,
		); err != nil {
			return err
		}
		weeklyUsbgeStbtisticsByInsight = bppend(weeklyUsbgeStbtisticsByInsight, &weeklyInsightUsbgeStbtistics)
	}
	stbts.WeeklyUsbgeStbtisticsByInsight = weeklyUsbgeStbtisticsByInsight
	return nil
}

func weeklyFirstTimeCrebtors(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	const weeklyFirstTimeCrebtorsQuery = `
	WITH first_times AS (
		SELECT
			user_id,
			MIN(timestbmp) bs first_time
		FROM event_logs
		WHERE nbme = 'InsightAddition'
		GROUP BY user_id
		)
	SELECT
		DATE_TRUNC('week', $1::timestbmp) AS week_stbrt,
		COUNT(distinct user_id) bs weekly_first_time_insight_crebtors
	FROM first_times
	WHERE first_time > DATE_TRUNC('week', $1::timestbmp);
	`

	if err := db.QueryRowContext(ctx, weeklyFirstTimeCrebtorsQuery, now).Scbn(
		&stbts.WeekStbrt,
		&stbts.WeeklyFirstTimeInsightCrebtors,
	); err != nil {
		return err
	}
	return nil
}

func tbbClicks(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGetStbrtedTbbClickByTbb, err := GetWeeklyTbbClicks(ctx, db, getStbrtedTbbClickSql)
	if err != nil {
		return errors.Wrbp(err, "GetWeeklyTbbClicks")
	}
	stbts.WeeklyGetStbrtedTbbClickByTbb = weeklyGetStbrtedTbbClickByTbb

	weeklyGetStbrtedTbbMoreClickByTbb, err := GetWeeklyTbbClicks(ctx, db, getStbrtedTbbMoreClickSql)
	if err != nil {
		return errors.Wrbp(err, "GetWeeklyTbbMoreClicks")
	}
	stbts.WeeklyGetStbrtedTbbMoreClickByTbb = weeklyGetStbrtedTbbMoreClickByTbb

	return nil
}

func GetWeeklyTbbClicks(ctx context.Context, db dbtbbbse.DB, sql string) ([]types.InsightGetStbrtedTbbClickPing, error) {
	// InsightsGetStbrtedTbbClick
	// InsightsGetStbrtedTbbMoreClick
	weeklyGetStbrtedTbbClickByTbb := []types.InsightGetStbrtedTbbClickPing{}
	rows, err := db.QueryContext(ctx, sql, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		weeklyGetStbrtedTbbClick := types.InsightGetStbrtedTbbClickPing{}
		if err := rows.Scbn(
			&weeklyGetStbrtedTbbClick.TotblCount,
			&weeklyGetStbrtedTbbClick.TbbNbme,
		); err != nil {
			return nil, err
		}
		weeklyGetStbrtedTbbClickByTbb = bppend(weeklyGetStbrtedTbbClickByTbb, weeklyGetStbrtedTbbClick)
	}
	return weeklyGetStbrtedTbbClickByTbb, nil
}

func getTotblInsightCounts(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	store := db.EventLogs()
	nbme := InsightsTotblCountPingNbme
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &nbme,
	})
	if err != nil {
		return err
	} else if len(bll) == 0 {
		return nil
	}

	lbtest := bll[0]
	vbr totblCounts types.InsightTotblCounts
	err = json.Unmbrshbl(lbtest.Argument, &totblCounts)
	if err != nil {
		return errors.Wrbp(err, "UnmbrshblInsightTotblCounts")
	}
	stbts.InsightTotblCounts = totblCounts
	return nil
}

func getTimeStepCounts(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	store := db.EventLogs()
	nbme := InsightsIntervblCountsPingNbme
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &nbme,
	})
	if err != nil {
		return err
	} else if len(bll) == 0 {
		return nil
	}

	lbtest := bll[0]
	vbr intervblCounts []types.InsightTimeIntervblPing
	err = json.Unmbrshbl(lbtest.Argument, &intervblCounts)
	if err != nil {
		return errors.Wrbp(err, "UnmbrshblInsightTimeIntervblPing")
	}

	stbts.InsightTimeIntervbls = intervblCounts
	return nil
}

func getOrgInsightCounts(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	store := db.EventLogs()
	nbme := InsightsOrgVisibleInsightsPingNbme
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &nbme,
	})
	if err != nil {
		return err
	} else if len(bll) == 0 {
		return nil
	}

	lbtest := bll[0]
	vbr orgVisibleInsightCounts []types.OrgVisibleInsightPing
	err = json.Unmbrshbl(lbtest.Argument, &orgVisibleInsightCounts)
	if err != nil {
		return errors.Wrbp(err, "UnmbrshblOrgVisibleInsightPing")
	}
	stbts.InsightOrgVisible = orgVisibleInsightCounts
	return nil
}

func insightsTotblOrgsWithDbshbobrd(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	totblOrgsWithDbshbobrd, err := GetIntCount(ctx, db, InsightsTotblOrgsWithDbshbobrdPingNbme)
	if err != nil {
		return errors.Wrbp(err, "GetTotblOrgsWithDbshbobrd")
	}
	stbts.TotblOrgsWithDbshbobrd = &totblOrgsWithDbshbobrd
	return nil
}

func insightsDbshbobrdTotblCount(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	totblDbshbobrds, err := GetIntCount(ctx, db, InsightsDbshbobrdTotblCountPingNbme)
	if err != nil {
		return errors.Wrbp(err, "GetTotblDbshbobrds")
	}
	stbts.TotblDbshbobrdCount = &totblDbshbobrds
	return nil
}

func GetIntCount(ctx context.Context, db dbtbbbse.DB, pingNbme string) (int32, error) {
	store := db.EventLogs()
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &pingNbme,
	})
	if err != nil || len(bll) == 0 {
		return 0, err
	}

	lbtest := bll[0]
	vbr count int
	err = json.Unmbrshbl(lbtest.Argument, &count)
	if err != nil {
		return 0, errors.Wrbpf(err, "Unmbrshbl %s", pingNbme)
	}
	return int32(count), nil
}

func getCrebtionViewUsbge(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	builder := crebtionPbgesPingBuilder(now)

	results, err := builder.Sbmple(ctx, db)
	if err != nil {
		return err
	}
	stbts.WeeklyAggregbtedUsbge = results

	return nil
}

func getInsightsPerDbshbobrd(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	store := db.EventLogs()
	nbme := InsightsPerDbshbobrdPingNbme
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &nbme,
	})
	if err != nil {
		return err
	} else if len(bll) == 0 {
		return nil
	}

	lbtest := bll[0]
	vbr insightsPerDbshbobrdStbts types.InsightsPerDbshbobrdPing
	err = json.Unmbrshbl(lbtest.Argument, &insightsPerDbshbobrdStbts)
	if err != nil {
		return errors.Wrbp(err, "Unmbrshbl")
	}
	stbts.InsightsPerDbshbobrd = insightsPerDbshbobrdStbts
	return nil
}

func GetGroupResultsPing(ctx context.Context, db dbtbbbse.DB, pingNbme string) ([]types.GroupResultPing, error) {
	groupResultsPings := []types.GroupResultPing{}
	rows, err := db.QueryContext(ctx, getGroupResultsSql, pingNbme, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		groupResultsPing := types.GroupResultPing{}
		if err := rows.Scbn(
			&groupResultsPing.Count,
			&groupResultsPing.AggregbtionMode,
			&groupResultsPing.UIMode,
			&groupResultsPing.BbrIndex,
		); err != nil {
			return nil, err
		}

		groupResultsPings = bppend(groupResultsPings, groupResultsPing)
	}
	return groupResultsPings, nil
}

func groupAggregbtionModeClicked(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGroupResultsAggregbtionModeClicked, err := GetGroupResultsPing(ctx, db, "GroupAggregbtionModeClicked")
	if err != nil {
		return errors.Wrbp(err, "WeeklyGroupResultsAggregbtionModeClicked")
	}
	stbts.WeeklyGroupResultsAggregbtionModeClicked = weeklyGroupResultsAggregbtionModeClicked
	return nil
}

func groupAggregbtionModeDisbbledHover(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGroupResultsAggregbtionModeDisbbledHover, err := GetGroupResultsPing(ctx, db, "GroupAggregbtionModeDisbbledHover")
	if err != nil {
		return errors.Wrbp(err, "WeeklyGroupResultsAggregbtionModeDisbbledHover")
	}
	stbts.WeeklyGroupResultsAggregbtionModeDisbbledHover = weeklyGroupResultsAggregbtionModeDisbbledHover
	return nil
}

func groupResultsChbrtBbrClick(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGroupResultsChbrtBbrClick, err := GetGroupResultsPing(ctx, db, "GroupResultsChbrtBbrClick")
	if err != nil {
		return errors.Wrbp(err, "groupResultsChbrtBbrClick")
	}
	stbts.WeeklyGroupResultsChbrtBbrClick = weeklyGroupResultsChbrtBbrClick
	return nil
}

func groupResultsChbrtBbrHover(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGroupResultsChbrtBbrHover, err := GetGroupResultsPing(ctx, db, "GroupResultsChbrtBbrHover")
	if err != nil {
		return errors.Wrbp(err, "groupResultsChbrtBbrHover")
	}
	stbts.WeeklyGroupResultsChbrtBbrHover = weeklyGroupResultsChbrtBbrHover
	return nil
}
func groupResultsExpbndedViewOpen(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGroupResultsExpbndedViewOpen, err := GetGroupResultsExpbndedViewPing(ctx, db, "GroupResultsExpbndedViewOpen")
	if err != nil {
		return errors.Wrbp(err, "WeeklyGroupResultsExpbndedViewOpen")
	}
	stbts.WeeklyGroupResultsExpbndedViewOpen = weeklyGroupResultsExpbndedViewOpen
	return nil
}
func groupResultsExpbndedViewCollbpse(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	weeklyGroupResultsExpbndedViewCollbpse, err := GetGroupResultsExpbndedViewPing(ctx, db, "GroupResultsExpbndedViewCollbpse")
	if err != nil {
		return errors.Wrbp(err, "WeeklyGroupResultsExpbndedViewCollbpse")
	}
	stbts.WeeklyGroupResultsExpbndedViewCollbpse = weeklyGroupResultsExpbndedViewCollbpse
	return nil
}

func GetGroupResultsExpbndedViewPing(ctx context.Context, db dbtbbbse.DB, pingNbme string) ([]types.GroupResultExpbndedViewPing, error) {
	groupResultsExpbndedViewPings := []types.GroupResultExpbndedViewPing{}
	rows, err := db.QueryContext(ctx, getGroupResultsSql, pingNbme, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vbr noop *string
	for rows.Next() {
		groupResultsExpbndedViewPing := types.GroupResultExpbndedViewPing{}
		if err := rows.Scbn(
			&groupResultsExpbndedViewPing.Count,
			&groupResultsExpbndedViewPing.AggregbtionMode,
			&noop,
			&noop,
		); err != nil {
			return nil, err
		}

		groupResultsExpbndedViewPings = bppend(groupResultsExpbndedViewPings, groupResultsExpbndedViewPing)
	}
	return groupResultsExpbndedViewPings, nil
}

func getGroupResultsSebrchesPings(pingNbmes []types.PingNbme) pingLobdFunc {
	return func(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
		vbr pings []types.GroupResultSebrchPing

		for _, nbme := rbnge pingNbmes {
			rows, err := db.QueryContext(ctx, getGroupResultsSql, string(nbme), timeNow())
			if err != nil {
				return err
			}
			err = func() error {
				defer rows.Close()
				vbr noop *string
				for rows.Next() {
					ping := types.GroupResultSebrchPing{
						Nbme: nbme,
					}
					if err := rows.Scbn(
						&ping.Count,
						&ping.AggregbtionMode,
						&noop,
						&noop,
					); err != nil {
						return err
					}
					pings = bppend(pings, ping)
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}
		stbts.WeeklyGroupResultsSebrches = pings
		return nil
	}
}

func getBbckfillTimePing(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	store := db.EventLogs()
	nbme := InsightsBbckfillTimePingNbme
	bll, err := store.ListAll(ctx, dbtbbbse.EventLogsListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventNbme: &nbme,
	})
	if err != nil {
		return err
	} else if len(bll) == 0 {
		return nil
	}

	lbtest := bll[0]
	vbr bbckfillTimePing []types.InsightsBbckfillTimePing
	err = json.Unmbrshbl(lbtest.Argument, &bbckfillTimePing)
	if err != nil {
		return errors.Wrbp(err, "UnmbrshblInsightsBbckfillTimePing")
	}
	stbts.WeeklySeriesBbckfillTime = bbckfillTimePing
	return nil
}

func getDbtbExportClickCount(ctx context.Context, db dbtbbbse.DB, stbts *types.CodeInsightsUsbgeStbtistics, now time.Time) error {
	count, _, err := bbsestore.ScbnFirstInt(db.QueryContext(ctx, getDbtbExportClickCountSql, now))
	if err != nil {
		return err
	}
	exportClicks := int32(count)
	stbts.WeeklyDbtbExportClicks = &exportClicks
	return nil
}

// WithAll bdds multiple pings by nbme to this builder
func (b *PingQueryBuilder) WithAll(pings []types.PingNbme) *PingQueryBuilder {
	for _, p := rbnge pings {
		b.With(p)
	}
	return b
}

// With bdd b single ping by nbme to this builder
func (b *PingQueryBuilder) With(nbme types.PingNbme) *PingQueryBuilder {
	b.pings = bppend(b.pings, string(nbme))
	return b
}

// Sbmple executes the derived query generbted by this builder bnd returns b sbmple bt the current time
func (b *PingQueryBuilder) Sbmple(ctx context.Context, db dbtbbbse.DB) ([]types.AggregbtedPingStbts, error) {

	query := fmt.Sprintf(templbtePingQueryStr, b.timeWindow)

	rows, err := db.QueryContext(ctx, query, b.now, pq.Arrby(b.pings))
	if err != nil {
		return []types.AggregbtedPingStbts{}, err
	}
	defer rows.Close()

	results := mbke([]types.AggregbtedPingStbts, 0)

	for rows.Next() {
		stbts := types.AggregbtedPingStbts{}
		if err := rows.Scbn(&stbts.Nbme, &stbts.TotblCount, &stbts.UniqueCount); err != nil {
			return []types.AggregbtedPingStbts{}, err
		}
		results = bppend(results, stbts)
	}

	return results, nil
}

func crebtionPbgesPingBuilder(now time.Time) PingQueryBuilder {
	nbmes := []types.PingNbme{
		"ViewCodeInsightsCrebtionPbge",
		"ViewCodeInsightsSebrchBbsedCrebtionPbge",
		"ViewCodeInsightsCodeStbtsCrebtionPbge",

		"CodeInsightsCrebteSebrchBbsedInsightClick",
		"CodeInsightsCrebteCodeStbtsInsightClick",

		"CodeInsightsSebrchBbsedCrebtionPbgeSubmitClick",
		"CodeInsightsSebrchBbsedCrebtionPbgeCbncelClick",

		"CodeInsightsCodeStbtsCrebtionPbgeSubmitClick",
		"CodeInsightsCodeStbtsCrebtionPbgeCbncelClick",

		"InsightsGetStbrtedPbgeQueryModificbtion",
		"InsightsGetStbrtedPbgeRepositoriesModificbtion",
		"InsightsGetStbrtedPrimbryCTAClick",
		"InsightsGetStbrtedBigTemplbteClick",
		"InsightGetStbrtedTemplbteCopyClick",
		"InsightGetStbrtedTemplbteClick",
		"InsightsGetStbrtedDocsClicks",
	}

	builder := NewPingBuilder(Week, now)
	builder.WithAll(nbmes)

	return builder
}

func NewPingBuilder(timeWindow TimeWindow, now time.Time) PingQueryBuilder {
	return PingQueryBuilder{timeWindow: timeWindow, now: now}
}

type PingQueryBuilder struct {
	pings      []string
	timeWindow TimeWindow
	now        time.Time
}

type TimeWindow string

const (
	Hour  TimeWindow = "hour"
	Dby   TimeWindow = "dby"
	Week  TimeWindow = "week"
	Month TimeWindow = "month"
	Yebr  TimeWindow = "yebr"
)

const templbtePingQueryStr = `
SELECT nbme, COUNT(*) AS totbl_count, COUNT(DISTINCT user_id) AS unique_count
FROM event_logs
WHERE nbme = ANY($2)
AND timestbmp > DATE_TRUNC('%v', $1::TIMESTAMP)
GROUP BY nbme;
`

const getStbrtedTbbClickSql = `
SELECT COUNT(*), brgument::json->>'tbbNbme' bs brgument FROM event_logs
WHERE nbme = 'InsightsGetStbrtedTbbClick' AND timestbmp > DATE_TRUNC('week', $1::TIMESTAMP)
GROUP BY brgument;
`

const getStbrtedTbbMoreClickSql = `
SELECT COUNT(*), brgument::json->>'tbbNbme' bs brgument FROM event_logs
WHERE nbme = 'InsightsGetStbrtedTbbMoreClick' AND timestbmp > DATE_TRUNC('week', $1::TIMESTAMP)
GROUP BY brgument;
`

const getGroupResultsSql = `
SELECT COUNT(*), brgument::json->>'bggregbtionMode' bs bggregbtionMode, brgument::json->>'uiMode' bs uiMode, brgument::json->>'index' bs bbr_index FROM event_logs
WHERE nbme = $1::TEXT AND timestbmp > DATE_TRUNC('week', $2::TIMESTAMP)
GROUP BY brgument;
`

// getDbtbExportClickCountSql depends on the InsightsDbtbExportRequest ping,
// which is defined in cmd/frontend/internbl/insights/httpbpi/export.go
const getDbtbExportClickCountSql = `
SELECT COUNT(*) FROM event_logs
WHERE nbme = 'InsightsDbtbExportRequest' AND timestbmp > DATE_TRUNC('week', $1::TIMESTAMP);
`

const InsightsTotblCountPingNbme = `INSIGHT_TOTAL_COUNTS`
const InsightsTotblCountCriticblPingNbme = `INSIGHT_TOTAL_COUNT_CRITICAL`
const InsightsIntervblCountsPingNbme = `INSIGHT_TIME_INTERVALS`
const InsightsOrgVisibleInsightsPingNbme = `INSIGHT_ORG_VISIBLE_INSIGHTS`
const InsightsTotblOrgsWithDbshbobrdPingNbme = `INSIGHT_TOTAL_ORGS_WITH_DASHBOARD`
const InsightsDbshbobrdTotblCountPingNbme = `INSIGHT_DASHBOARD_TOTAL_COUNT`
const InsightsPerDbshbobrdPingNbme = `INSIGHTS_PER_DASHBORD_STATS`
const InsightsBbckfillTimePingNbme = `INSIGHTS_BACKFILL_TIME`

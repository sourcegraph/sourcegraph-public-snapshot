package usagestats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetCodeInsightsUsageStatistics(ctx context.Context, db database.DB) (*types.CodeInsightsUsageStatistics, error) {
	stats := types.CodeInsightsUsageStatistics{}

	const platformQuery = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'ViewInsights')                       			AS weekly_insights_page_views,
		COUNT(*) FILTER (WHERE name = 'ViewInsightsGetStartedPage')         			AS weekly_insights_get_started_page_views,
		COUNT(*) FILTER (WHERE name = 'StandaloneInsightPageViewed')					AS weekly_standalone_insight_page_views,
		COUNT(*) FILTER (WHERE name = 'StandaloneInsightDashboardClick') 				AS weekly_standalone_dashboard_clicks,
        COUNT(*) FILTER (WHERE name = 'StandaloneInsightPageEditClick') 				AS weekly_standalone_edit_clicks,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsights')        			AS weekly_insights_unique_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsightsGetStartedPage')  	AS weekly_insights_get_started_unique_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'StandaloneInsightPageViewed') 	AS weekly_standalone_insight_unique_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'StandaloneInsightDashboardClick') AS weekly_standalone_insight_unique_dashboard_clicks,
		COUNT(distinct user_id) FILTER (WHERE name = 'StandaloneInsightPageEditClick')  AS weekly_standalone_insight_unique_edit_clicks,
		COUNT(distinct user_id) FILTER (WHERE name = 'InsightAddition')					AS weekly_insight_creators,
		COUNT(*) FILTER (WHERE name = 'InsightConfigureClick') 							AS weekly_insight_configure_click,
		COUNT(*) FILTER (WHERE name = 'InsightAddMoreClick') 							AS weekly_insight_add_more_click,
		COUNT(*) FILTER (WHERE name = 'GroupResultsOpenSection') 						AS weekly_group_results_open_section,
		COUNT(*) FILTER (WHERE name = 'GroupResultsCollapseSection') 					AS weekly_group_results_collapse_section,
		COUNT(*) FILTER (WHERE name = 'GroupResultsInfoIconHover') 						AS weekly_group_results_info_icon_hover
	FROM event_logs
	WHERE name in ('ViewInsights', 'StandaloneInsightPageViewed', 'StandaloneInsightDashboardClick', 'StandaloneInsightPageEditClick',
			'ViewInsightsGetStartedPage', 'InsightAddition', 'InsightConfigureClick', 'InsightAddMoreClick', 'GroupResultsOpenSection',
			'GroupResultsCollapseSection', 'GroupResultsInfoIconHover')
		AND timestamp > DATE_TRUNC('week', $1::timestamp);
	`

	if err := db.QueryRowContext(ctx, platformQuery, timeNow()).Scan(
		&stats.WeeklyInsightsPageViews,
		&stats.WeeklyInsightsGetStartedPageViews,
		&stats.WeeklyStandaloneInsightPageViews,
		&stats.WeeklyStandaloneDashboardClicks,
		&stats.WeeklyStandaloneEditClicks,
		&stats.WeeklyInsightsUniquePageViews,
		&stats.WeeklyInsightsGetStartedUniquePageViews,
		&stats.WeeklyStandaloneInsightUniquePageViews,
		&stats.WeeklyStandaloneInsightUniqueDashboardClicks,
		&stats.WeeklyStandaloneInsightUniqueEditClicks,
		&stats.WeeklyInsightCreators,
		&stats.WeeklyInsightConfigureClick,
		&stats.WeeklyInsightAddMoreClick,
		&stats.WeeklyGroupResultsOpenSection,
		&stats.WeeklyGroupResultsCollapseSection,
		&stats.WeeklyGroupResultsInfoIconHover,
	); err != nil {
		return nil, err
	}

	const metricsByInsightQuery = `
	SELECT argument ->> 'insightType'::text 					             		AS insight_type,
        COUNT(*) FILTER (WHERE name = 'InsightAddition') 		             		AS additions,
        COUNT(*) FILTER (WHERE name = 'InsightEdit') 			             		AS edits,
        COUNT(*) FILTER (WHERE name = 'InsightRemoval') 		             		AS removals,
		COUNT(*) FILTER (WHERE name = 'InsightHover') 			             		AS hovers,
		COUNT(*) FILTER (WHERE name = 'InsightUICustomization') 			 		AS ui_customizations,
		COUNT(*) FILTER (WHERE name = 'InsightDataPointClick') 				 		AS data_point_clicks,
		COUNT(*) FILTER (WHERE name = 'InsightFiltersChange') 				 		AS filters_change
	FROM event_logs
	WHERE name in ('InsightAddition', 'InsightEdit', 'InsightRemoval', 'InsightHover', 'InsightUICustomization', 'InsightDataPointClick', 'InsightFiltersChange')
		AND timestamp > DATE_TRUNC('week', $1::timestamp)
	GROUP BY insight_type;
	`

	weeklyUsageStatisticsByInsight := []*types.InsightUsageStatistics{}
	rows, err := db.QueryContext(ctx, metricsByInsightQuery, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		weeklyInsightUsageStatistics := types.InsightUsageStatistics{}

		if err := rows.Scan(
			&weeklyInsightUsageStatistics.InsightType,
			&weeklyInsightUsageStatistics.Additions,
			&weeklyInsightUsageStatistics.Edits,
			&weeklyInsightUsageStatistics.Removals,
			&weeklyInsightUsageStatistics.Hovers,
			&weeklyInsightUsageStatistics.UICustomizations,
			&weeklyInsightUsageStatistics.DataPointClicks,
			&weeklyInsightUsageStatistics.FiltersChange,
		); err != nil {
			return nil, err
		}

		weeklyUsageStatisticsByInsight = append(weeklyUsageStatisticsByInsight, &weeklyInsightUsageStatistics)
	}
	stats.WeeklyUsageStatisticsByInsight = weeklyUsageStatisticsByInsight

	if err := rows.Err(); err != nil {
		return nil, err
	}

	const weeklyFirstTimeCreatorsQuery = `
	WITH first_times AS (
		SELECT
			user_id,
			MIN(timestamp) as first_time
		FROM event_logs
		WHERE name = 'InsightAddition'
		GROUP BY user_id
		)
	SELECT
		DATE_TRUNC('week', $1::timestamp) AS week_start,
		COUNT(distinct user_id) as weekly_first_time_insight_creators
	FROM first_times
	WHERE first_time > DATE_TRUNC('week', $1::timestamp);
	`

	if err := db.QueryRowContext(ctx, weeklyFirstTimeCreatorsQuery, timeNow()).Scan(
		&stats.WeekStart,
		&stats.WeeklyFirstTimeInsightCreators,
	); err != nil {
		return nil, err
	}

	weeklyUsage, err := GetCreationViewUsage(ctx, db, timeNow)
	if err != nil {
		return nil, err
	}
	stats.WeeklyAggregatedUsage = weeklyUsage

	// These two pings are slightly more fragile than the others because they deserialize json from settings. They are
	// also less important. So, in the case of any errors here we will not fail the entire ping for code insights.
	timeIntervals, err := GetTimeStepCounts(ctx, db)
	if err != nil {
		log15.Error("code-insights/GetTimeStepCounts", "error", err)
		return nil, nil
	}
	stats.InsightTimeIntervals = timeIntervals

	orgVisible, err := GetOrgInsightCounts(ctx, db)
	if err != nil {
		log15.Error("code-insights/GetOrgInsightCounts", "error", err)
		return nil, nil
	}
	stats.InsightOrgVisible = orgVisible

	totalCounts, err := GetTotalInsightCounts(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "GetTotalInsightCounts")
	}
	stats.InsightTotalCounts = totalCounts

	weeklyGetStartedTabClickByTab, err := GetWeeklyTabClicks(ctx, db, getStartedTabClickSql)
	if err != nil {
		return nil, errors.Wrap(err, "GetWeeklyTabClicks")
	}
	stats.WeeklyGetStartedTabClickByTab = weeklyGetStartedTabClickByTab

	weeklyGetStartedTabMoreClickByTab, err := GetWeeklyTabClicks(ctx, db, getStartedTabMoreClickSql)
	if err != nil {
		return nil, errors.Wrap(err, "GetWeeklyTabMoreClicks")
	}
	stats.WeeklyGetStartedTabMoreClickByTab = weeklyGetStartedTabMoreClickByTab

	totalOrgsWithDashboard, err := GetIntCount(ctx, db, InsightsTotalOrgsWithDashboardPingName)
	if err != nil {
		return nil, errors.Wrap(err, "GetTotalOrgsWithDashboard")
	}
	stats.TotalOrgsWithDashboard = &totalOrgsWithDashboard

	totalDashboards, err := GetIntCount(ctx, db, InsightsDashboardTotalCountPingName)
	if err != nil {
		return nil, errors.Wrap(err, "GetTotalDashboards")
	}
	stats.TotalDashboardCount = &totalDashboards

	insightsPerDashboard, err := GetInsightsPerDashboard(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "GetInsightsPerDashboard")
	}
	stats.InsightsPerDashboard = insightsPerDashboard

	weeklyGroupResultsAggregationModeClicked, err := GetGroupResultsPing(ctx, db, "GroupAggregationModeClicked")
	if err != nil {
		return nil, errors.Wrap(err, "WeeklyGroupResultsAggregationModeClicked")
	}
	stats.WeeklyGroupResultsAggregationModeClicked = weeklyGroupResultsAggregationModeClicked

	weeklyGroupResultsAggregationModeDisabledHover, err := GetGroupResultsPing(ctx, db, "GroupAggregationModeDisabledHover")
	if err != nil {
		return nil, errors.Wrap(err, "WeeklyGroupResultsAggregationModeDisabledHover")
	}
	stats.WeeklyGroupResultsAggregationModeDisabledHover = weeklyGroupResultsAggregationModeDisabledHover

	weeklyGroupResultsChartBarClick, err := GetGroupResultsPing(ctx, db, "GroupResultsChartBarClick")
	if err != nil {
		return nil, errors.Wrap(err, "GroupResultsChartBarClick")
	}
	stats.WeeklyGroupResultsChartBarClick = weeklyGroupResultsChartBarClick

	weeklyGroupResultsChartBarHover, err := GetGroupResultsPing(ctx, db, "GroupResultsChartBarHover")
	if err != nil {
		return nil, errors.Wrap(err, "GroupResultsChartBarHover")
	}
	stats.WeeklyGroupResultsChartBarHover = weeklyGroupResultsChartBarHover

	weeklyGroupResultsExpandedViewOpen, err := GetGroupResultsExpandedViewPing(ctx, db, "GroupResultsExpandedViewOpen")
	if err != nil {
		return nil, errors.Wrap(err, "WeeklyGroupResultsExpandedViewOpen")
	}
	stats.WeeklyGroupResultsExpandedViewOpen = weeklyGroupResultsExpandedViewOpen

	weeklyGroupResultsExpandedViewCollapse, err := GetGroupResultsExpandedViewPing(ctx, db, "GroupResultsExpandedViewCollapse")
	if err != nil {
		return nil, errors.Wrap(err, "WeeklyGroupResultsExpandedViewCollapse")
	}
	stats.WeeklyGroupResultsExpandedViewCollapse = weeklyGroupResultsExpandedViewCollapse

	return &stats, nil
}

func GetWeeklyTabClicks(ctx context.Context, db database.DB, sql string) ([]types.InsightGetStartedTabClickPing, error) {
	weeklyGetStartedTabClickByTab := []types.InsightGetStartedTabClickPing{}
	rows, err := db.QueryContext(ctx, sql, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		weeklyGetStartedTabClick := types.InsightGetStartedTabClickPing{}
		if err := rows.Scan(
			&weeklyGetStartedTabClick.TotalCount,
			&weeklyGetStartedTabClick.TabName,
		); err != nil {
			return nil, err
		}
		weeklyGetStartedTabClickByTab = append(weeklyGetStartedTabClickByTab, weeklyGetStartedTabClick)
	}
	return weeklyGetStartedTabClickByTab, nil
}

func GetTotalInsightCounts(ctx context.Context, db database.DB) (types.InsightTotalCounts, error) {
	store := db.EventLogs()
	name := InsightsTotalCountPingName
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &name,
	})
	if err != nil {
		return types.InsightTotalCounts{}, err
	} else if len(all) == 0 {
		return types.InsightTotalCounts{}, nil
	}

	latest := all[0]
	var totalCounts types.InsightTotalCounts
	err = json.Unmarshal([]byte(latest.Argument), &totalCounts)
	if err != nil {
		return types.InsightTotalCounts{}, errors.Wrap(err, "Unmarshal")
	}
	return totalCounts, err
}

func GetTimeStepCounts(ctx context.Context, db database.DB) ([]types.InsightTimeIntervalPing, error) {
	store := db.EventLogs()
	name := InsightsIntervalCountsPingName
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &name,
	})
	if err != nil {
		return []types.InsightTimeIntervalPing{}, err
	} else if len(all) == 0 {
		return []types.InsightTimeIntervalPing{}, nil
	}

	latest := all[0]
	var intervalCounts []types.InsightTimeIntervalPing
	err = json.Unmarshal([]byte(latest.Argument), &intervalCounts)
	if err != nil {
		return []types.InsightTimeIntervalPing{}, errors.Wrap(err, "Unmarshal")
	}
	return intervalCounts, nil
}

func GetOrgInsightCounts(ctx context.Context, db database.DB) ([]types.OrgVisibleInsightPing, error) {
	store := db.EventLogs()
	name := InsightsOrgVisibleInsightsPingName
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &name,
	})
	if err != nil {
		return []types.OrgVisibleInsightPing{}, err
	} else if len(all) == 0 {
		return []types.OrgVisibleInsightPing{}, nil
	}

	latest := all[0]
	var orgVisibleInsightCounts []types.OrgVisibleInsightPing
	err = json.Unmarshal([]byte(latest.Argument), &orgVisibleInsightCounts)
	if err != nil {
		return []types.OrgVisibleInsightPing{}, errors.Wrap(err, "Unmarshal")
	}
	return orgVisibleInsightCounts, nil
}

func GetIntCount(ctx context.Context, db database.DB, pingName string) (int32, error) {
	store := db.EventLogs()
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &pingName,
	})
	if err != nil || len(all) == 0 {
		return 0, err
	}

	latest := all[0]
	var count int
	err = json.Unmarshal([]byte(latest.Argument), &count)
	if err != nil {
		return 0, errors.Wrap(err, "Unmarshal")
	}
	return int32(count), nil
}

func GetCreationViewUsage(ctx context.Context, db database.DB, timeSupplier func() time.Time) ([]types.AggregatedPingStats, error) {
	builder := creationPagesPingBuilder(timeSupplier)

	results, err := builder.Sample(ctx, db)
	if err != nil {
		return []types.AggregatedPingStats{}, err
	}

	return results, nil
}

func GetInsightsPerDashboard(ctx context.Context, db database.DB) (types.InsightsPerDashboardPing, error) {
	store := db.EventLogs()
	name := InsightsPerDashboardPingName
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &name,
	})
	if err != nil {
		return types.InsightsPerDashboardPing{}, err
	} else if len(all) == 0 {
		return types.InsightsPerDashboardPing{}, nil
	}

	latest := all[0]
	var insightsPerDashboardStats types.InsightsPerDashboardPing
	err = json.Unmarshal([]byte(latest.Argument), &insightsPerDashboardStats)
	if err != nil {
		return types.InsightsPerDashboardPing{}, errors.Wrap(err, "Unmarshal")
	}
	return insightsPerDashboardStats, nil
}

func GetGroupResultsPing(ctx context.Context, db database.DB, pingName string) ([]types.GroupResultPing, error) {
	groupResultsPings := []types.GroupResultPing{}
	rows, err := db.QueryContext(ctx, getGroupResultsSql, pingName, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		groupResultsPing := types.GroupResultPing{}
		if err := rows.Scan(
			&groupResultsPing.Count,
			&groupResultsPing.AggregationMode,
			&groupResultsPing.UIMode,
		); err != nil {
			return nil, err
		}

		groupResultsPings = append(groupResultsPings, groupResultsPing)
	}
	return groupResultsPings, nil
}

func GetGroupResultsExpandedViewPing(ctx context.Context, db database.DB, pingName string) ([]types.GroupResultExpandedViewPing, error) {
	groupResultsExpandedViewPings := []types.GroupResultExpandedViewPing{}
	rows, err := db.QueryContext(ctx, getGroupResultsSql, pingName, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var noop *string
	for rows.Next() {
		groupResultsExpandedViewPing := types.GroupResultExpandedViewPing{}
		if err := rows.Scan(
			&groupResultsExpandedViewPing.Count,
			&groupResultsExpandedViewPing.AggregationMode,
			&noop,
		); err != nil {
			return nil, err
		}

		groupResultsExpandedViewPings = append(groupResultsExpandedViewPings, groupResultsExpandedViewPing)
	}
	return groupResultsExpandedViewPings, nil
}

// WithAll adds multiple pings by name to this builder
func (b *PingQueryBuilder) WithAll(pings []types.PingName) *PingQueryBuilder {
	for _, p := range pings {
		b.With(p)
	}
	return b
}

// With add a single ping by name to this builder
func (b *PingQueryBuilder) With(name types.PingName) *PingQueryBuilder {
	b.pings = append(b.pings, string(name))
	return b
}

// Sample executes the derived query generated by this builder and returns a sample at the current time
func (b *PingQueryBuilder) Sample(ctx context.Context, db database.DB) ([]types.AggregatedPingStats, error) {

	query := fmt.Sprintf(templatePingQueryStr, b.timeWindow)

	rows, err := db.QueryContext(ctx, query, b.getTime(), pq.Array(b.pings))
	if err != nil {
		return []types.AggregatedPingStats{}, err
	}
	defer rows.Close()

	results := make([]types.AggregatedPingStats, 0)

	for rows.Next() {
		stats := types.AggregatedPingStats{}
		if err := rows.Scan(&stats.Name, &stats.TotalCount, &stats.UniqueCount); err != nil {
			return []types.AggregatedPingStats{}, err
		}
		results = append(results, stats)
	}

	return results, nil
}

func creationPagesPingBuilder(timeSupplier func() time.Time) PingQueryBuilder {
	names := []types.PingName{
		"ViewCodeInsightsCreationPage",
		"ViewCodeInsightsSearchBasedCreationPage",
		"ViewCodeInsightsCodeStatsCreationPage",

		"CodeInsightsCreateSearchBasedInsightClick",
		"CodeInsightsCreateCodeStatsInsightClick",
		"CodeInsightsExploreInsightExtensionsClick",

		"CodeInsightsSearchBasedCreationPageSubmitClick",
		"CodeInsightsSearchBasedCreationPageCancelClick",

		"CodeInsightsCodeStatsCreationPageSubmitClick",
		"CodeInsightsCodeStatsCreationPageCancelClick",

		"InsightsGetStartedPageQueryModification",
		"InsightsGetStartedPageRepositoriesModification",
		"InsightsGetStartedPrimaryCTAClick",
		"InsightsGetStartedBigTemplateClick",
		"InsightGetStartedTemplateCopyClick",
		"InsightGetStartedTemplateClick",
		"InsightsGetStartedDocsClicks",
	}

	builder := NewPingBuilder(Week, timeSupplier)
	builder.WithAll(names)

	return builder
}

func NewPingBuilder(timeWindow TimeWindow, timeSupplier func() time.Time) PingQueryBuilder {
	return PingQueryBuilder{timeWindow: timeWindow, getTime: timeSupplier}
}

type PingQueryBuilder struct {
	pings      []string
	timeWindow TimeWindow
	getTime    func() time.Time
}

type TimeWindow string

const (
	Hour  TimeWindow = "hour"
	Day   TimeWindow = "day"
	Week  TimeWindow = "week"
	Month TimeWindow = "month"
	Year  TimeWindow = "year"
)

const templatePingQueryStr = `
-- source:internal/usagestats/code_insights.go:Sample
SELECT name, COUNT(*) AS total_count, COUNT(DISTINCT user_id) AS unique_count
FROM event_logs
WHERE name = ANY($2)
AND timestamp > DATE_TRUNC('%v', $1::TIMESTAMP)
GROUP BY name;
`

const getStartedTabClickSql = `
SELECT COUNT(*), argument::json->>'tabName' as argument FROM event_logs
WHERE name = 'InsightsGetStartedTabClick' AND timestamp > DATE_TRUNC('week', $1::TIMESTAMP)
GROUP BY argument;
`

const getStartedTabMoreClickSql = `
SELECT COUNT(*), argument::json->>'tabName' as argument FROM event_logs
WHERE name = 'InsightsGetStartedTabMoreClick' AND timestamp > DATE_TRUNC('week', $1::TIMESTAMP)
GROUP BY argument;
`

const getGroupResultsSql = `
SELECT COUNT(*), argument::json->>'aggregationMode' as aggregationMode, argument::json->>'uiMode' as uiMode FROM event_logs
WHERE name = $1::TEXT AND timestamp > DATE_TRUNC('week', $2::TIMESTAMP)
GROUP BY argument;
`

const InsightsTotalCountPingName = `INSIGHT_TOTAL_COUNTS`
const InsightsTotalCountCriticalPingName = `INSIGHT_TOTAL_COUNT_CRITICAL`
const InsightsIntervalCountsPingName = `INSIGHT_TIME_INTERVALS`
const InsightsOrgVisibleInsightsPingName = `INSIGHT_ORG_VISIBLE_INSIGHTS`
const InsightsTotalOrgsWithDashboardPingName = `INSIGHT_TOTAL_ORGS_WITH_DASHBOARD`
const InsightsDashboardTotalCountPingName = `INSIGHT_DASHBOARD_TOTAL_COUNT`
const InsightsPerDashboardPingName = `INSIGHTS_PER_DASHBORD_STATS`

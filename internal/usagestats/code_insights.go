package usagestats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sourcegraph/log"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type pingLoadFunc func(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error

type pingLoader struct {
	now        time.Time
	operations map[string]pingLoadFunc
}

func newPingLoader(now time.Time) *pingLoader {
	return &pingLoader{now: now, operations: make(map[string]pingLoadFunc)}
}

func (p *pingLoader) withOperation(name string, loadFunc pingLoadFunc) {
	p.operations[name] = loadFunc
}

func (p *pingLoader) generate(ctx context.Context, db database.DB) *types.CodeInsightsUsageStatistics {
	stats := &types.CodeInsightsUsageStatistics{}
	logger := log.Scoped("code insights ping loader")

	for name, loadFunc := range p.operations {
		err := loadFunc(ctx, db, stats, p.now)
		if err != nil {
			logger.Error("insights pings loading error, skipping ping", log.String("name", name), log.Error(err))
		}
	}
	return stats
}

func GetCodeInsightsUsageStatistics(ctx context.Context, db database.DB) (*types.CodeInsightsUsageStatistics, error) {
	loader := newPingLoader(timeNow())

	loader.withOperation("weeklyUsage", weeklyUsage)
	loader.withOperation("weeklyMetricsByInsight", weeklyMetricsByInsight)
	loader.withOperation("weeklyFirstTimeCreators", weeklyFirstTimeCreators)
	loader.withOperation("getCreationViewUsage", getCreationViewUsage)
	loader.withOperation("getTimeStepCounts", getTimeStepCounts)
	loader.withOperation("getOrgInsightCounts", getOrgInsightCounts)
	loader.withOperation("getTotalInsightCounts", getTotalInsightCounts)
	loader.withOperation("tabClicks", tabClicks)
	loader.withOperation("insightsTotalOrgsWithDashboard", insightsTotalOrgsWithDashboard)
	loader.withOperation("insightsDashboardTotalCount", insightsDashboardTotalCount)
	loader.withOperation("getInsightsPerDashboard", getInsightsPerDashboard)

	loader.withOperation("groupAggregationModeClicked", groupAggregationModeClicked)
	loader.withOperation("groupAggregationModeDisabledHover", groupAggregationModeDisabledHover)
	loader.withOperation("groupResultsChartBarClick", groupResultsChartBarClick)
	loader.withOperation("groupResultsChartBarHover", groupResultsChartBarHover)
	loader.withOperation("groupResultsExpandedViewOpen", groupResultsExpandedViewOpen)
	loader.withOperation("groupResultsExpandedViewCollapse", groupResultsExpandedViewCollapse)
	loader.withOperation("getBackfillTimePing", getBackfillTimePing)
	loader.withOperation("getDataExportClicks", getDataExportClickCount)

	loader.withOperation("getGroupResultsSearchesPings", getGroupResultsSearchesPings(
		[]types.PingName{
			"ProactiveLimitHit",
			"ProactiveLimitSuccess",
			"ExplicitLimitHit",
			"ExplicitLimitSuccess",
		}))

	return loader.generate(ctx, db), nil
}

func weeklyUsage(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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
		return err
	}
	return nil
}

func weeklyMetricsByInsight(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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

	var weeklyUsageStatisticsByInsight []*types.InsightUsageStatistics
	rows, err := db.QueryContext(ctx, metricsByInsightQuery, timeNow())
	if err != nil {
		return err
	}

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
			return err
		}
		weeklyUsageStatisticsByInsight = append(weeklyUsageStatisticsByInsight, &weeklyInsightUsageStatistics)
	}
	stats.WeeklyUsageStatisticsByInsight = weeklyUsageStatisticsByInsight
	return nil
}

func weeklyFirstTimeCreators(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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

	if err := db.QueryRowContext(ctx, weeklyFirstTimeCreatorsQuery, now).Scan(
		&stats.WeekStart,
		&stats.WeeklyFirstTimeInsightCreators,
	); err != nil {
		return err
	}
	return nil
}

func tabClicks(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGetStartedTabClickByTab, err := GetWeeklyTabClicks(ctx, db, getStartedTabClickSql)
	if err != nil {
		return errors.Wrap(err, "GetWeeklyTabClicks")
	}
	stats.WeeklyGetStartedTabClickByTab = weeklyGetStartedTabClickByTab

	weeklyGetStartedTabMoreClickByTab, err := GetWeeklyTabClicks(ctx, db, getStartedTabMoreClickSql)
	if err != nil {
		return errors.Wrap(err, "GetWeeklyTabMoreClicks")
	}
	stats.WeeklyGetStartedTabMoreClickByTab = weeklyGetStartedTabMoreClickByTab

	return nil
}

func GetWeeklyTabClicks(ctx context.Context, db database.DB, sql string) ([]types.InsightGetStartedTabClickPing, error) {
	// InsightsGetStartedTabClick
	// InsightsGetStartedTabMoreClick
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

func getTotalInsightCounts(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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
		return err
	} else if len(all) == 0 {
		return nil
	}

	latest := all[0]
	var totalCounts types.InsightTotalCounts
	err = json.Unmarshal(latest.Argument, &totalCounts)
	if err != nil {
		return errors.Wrap(err, "UnmarshalInsightTotalCounts")
	}
	stats.InsightTotalCounts = totalCounts
	return nil
}

func getTimeStepCounts(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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
		return err
	} else if len(all) == 0 {
		return nil
	}

	latest := all[0]
	var intervalCounts []types.InsightTimeIntervalPing
	err = json.Unmarshal(latest.Argument, &intervalCounts)
	if err != nil {
		return errors.Wrap(err, "UnmarshalInsightTimeIntervalPing")
	}

	stats.InsightTimeIntervals = intervalCounts
	return nil
}

func getOrgInsightCounts(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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
		return err
	} else if len(all) == 0 {
		return nil
	}

	latest := all[0]
	var orgVisibleInsightCounts []types.OrgVisibleInsightPing
	err = json.Unmarshal(latest.Argument, &orgVisibleInsightCounts)
	if err != nil {
		return errors.Wrap(err, "UnmarshalOrgVisibleInsightPing")
	}
	stats.InsightOrgVisible = orgVisibleInsightCounts
	return nil
}

func insightsTotalOrgsWithDashboard(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	totalOrgsWithDashboard, err := GetIntCount(ctx, db, InsightsTotalOrgsWithDashboardPingName)
	if err != nil {
		return errors.Wrap(err, "GetTotalOrgsWithDashboard")
	}
	stats.TotalOrgsWithDashboard = &totalOrgsWithDashboard
	return nil
}

func insightsDashboardTotalCount(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	totalDashboards, err := GetIntCount(ctx, db, InsightsDashboardTotalCountPingName)
	if err != nil {
		return errors.Wrap(err, "GetTotalDashboards")
	}
	stats.TotalDashboardCount = &totalDashboards
	return nil
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
	err = json.Unmarshal(latest.Argument, &count)
	if err != nil {
		return 0, errors.Wrapf(err, "Unmarshal %s", pingName)
	}
	return int32(count), nil
}

func getCreationViewUsage(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	builder := creationPagesPingBuilder(now)

	results, err := builder.Sample(ctx, db)
	if err != nil {
		return err
	}
	stats.WeeklyAggregatedUsage = results

	return nil
}

func getInsightsPerDashboard(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
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
		return err
	} else if len(all) == 0 {
		return nil
	}

	latest := all[0]
	var insightsPerDashboardStats types.InsightsPerDashboardPing
	err = json.Unmarshal(latest.Argument, &insightsPerDashboardStats)
	if err != nil {
		return errors.Wrap(err, "Unmarshal")
	}
	stats.InsightsPerDashboard = insightsPerDashboardStats
	return nil
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
			&groupResultsPing.BarIndex,
		); err != nil {
			return nil, err
		}

		groupResultsPings = append(groupResultsPings, groupResultsPing)
	}
	return groupResultsPings, nil
}

func groupAggregationModeClicked(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGroupResultsAggregationModeClicked, err := GetGroupResultsPing(ctx, db, "GroupAggregationModeClicked")
	if err != nil {
		return errors.Wrap(err, "WeeklyGroupResultsAggregationModeClicked")
	}
	stats.WeeklyGroupResultsAggregationModeClicked = weeklyGroupResultsAggregationModeClicked
	return nil
}

func groupAggregationModeDisabledHover(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGroupResultsAggregationModeDisabledHover, err := GetGroupResultsPing(ctx, db, "GroupAggregationModeDisabledHover")
	if err != nil {
		return errors.Wrap(err, "WeeklyGroupResultsAggregationModeDisabledHover")
	}
	stats.WeeklyGroupResultsAggregationModeDisabledHover = weeklyGroupResultsAggregationModeDisabledHover
	return nil
}

func groupResultsChartBarClick(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGroupResultsChartBarClick, err := GetGroupResultsPing(ctx, db, "GroupResultsChartBarClick")
	if err != nil {
		return errors.Wrap(err, "groupResultsChartBarClick")
	}
	stats.WeeklyGroupResultsChartBarClick = weeklyGroupResultsChartBarClick
	return nil
}

func groupResultsChartBarHover(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGroupResultsChartBarHover, err := GetGroupResultsPing(ctx, db, "GroupResultsChartBarHover")
	if err != nil {
		return errors.Wrap(err, "groupResultsChartBarHover")
	}
	stats.WeeklyGroupResultsChartBarHover = weeklyGroupResultsChartBarHover
	return nil
}
func groupResultsExpandedViewOpen(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGroupResultsExpandedViewOpen, err := GetGroupResultsExpandedViewPing(ctx, db, "GroupResultsExpandedViewOpen")
	if err != nil {
		return errors.Wrap(err, "WeeklyGroupResultsExpandedViewOpen")
	}
	stats.WeeklyGroupResultsExpandedViewOpen = weeklyGroupResultsExpandedViewOpen
	return nil
}
func groupResultsExpandedViewCollapse(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	weeklyGroupResultsExpandedViewCollapse, err := GetGroupResultsExpandedViewPing(ctx, db, "GroupResultsExpandedViewCollapse")
	if err != nil {
		return errors.Wrap(err, "WeeklyGroupResultsExpandedViewCollapse")
	}
	stats.WeeklyGroupResultsExpandedViewCollapse = weeklyGroupResultsExpandedViewCollapse
	return nil
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
			&noop,
		); err != nil {
			return nil, err
		}

		groupResultsExpandedViewPings = append(groupResultsExpandedViewPings, groupResultsExpandedViewPing)
	}
	return groupResultsExpandedViewPings, nil
}

func getGroupResultsSearchesPings(pingNames []types.PingName) pingLoadFunc {
	return func(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
		var pings []types.GroupResultSearchPing

		for _, name := range pingNames {
			rows, err := db.QueryContext(ctx, getGroupResultsSql, string(name), timeNow())
			if err != nil {
				return err
			}
			err = func() error {
				defer rows.Close()
				var noop *string
				for rows.Next() {
					ping := types.GroupResultSearchPing{
						Name: name,
					}
					if err := rows.Scan(
						&ping.Count,
						&ping.AggregationMode,
						&noop,
						&noop,
					); err != nil {
						return err
					}
					pings = append(pings, ping)
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}
		stats.WeeklyGroupResultsSearches = pings
		return nil
	}
}

func getBackfillTimePing(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	store := db.EventLogs()
	name := InsightsBackfillTimePingName
	all, err := store.ListAll(ctx, database.EventLogsListOptions{
		LimitOffset: &database.LimitOffset{
			Limit:  1,
			Offset: 0,
		},
		EventName: &name,
	})
	if err != nil {
		return err
	} else if len(all) == 0 {
		return nil
	}

	latest := all[0]
	var backfillTimePing []types.InsightsBackfillTimePing
	err = json.Unmarshal(latest.Argument, &backfillTimePing)
	if err != nil {
		return errors.Wrap(err, "UnmarshalInsightsBackfillTimePing")
	}
	stats.WeeklySeriesBackfillTime = backfillTimePing
	return nil
}

func getDataExportClickCount(ctx context.Context, db database.DB, stats *types.CodeInsightsUsageStatistics, now time.Time) error {
	count, _, err := basestore.ScanFirstInt(db.QueryContext(ctx, getDataExportClickCountSql, now))
	if err != nil {
		return err
	}
	exportClicks := int32(count)
	stats.WeeklyDataExportClicks = &exportClicks
	return nil
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

	rows, err := db.QueryContext(ctx, query, b.now, pq.Array(b.pings))
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

func creationPagesPingBuilder(now time.Time) PingQueryBuilder {
	names := []types.PingName{
		"ViewCodeInsightsCreationPage",
		"ViewCodeInsightsSearchBasedCreationPage",
		"ViewCodeInsightsCodeStatsCreationPage",

		"CodeInsightsCreateSearchBasedInsightClick",
		"CodeInsightsCreateCodeStatsInsightClick",

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

	builder := NewPingBuilder(Week, now)
	builder.WithAll(names)

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
	Day   TimeWindow = "day"
	Week  TimeWindow = "week"
	Month TimeWindow = "month"
	Year  TimeWindow = "year"
)

const templatePingQueryStr = `
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
SELECT COUNT(*), argument::json->>'aggregationMode' as aggregationMode, argument::json->>'uiMode' as uiMode, argument::json->>'index' as bar_index FROM event_logs
WHERE name = $1::TEXT AND timestamp > DATE_TRUNC('week', $2::TIMESTAMP)
GROUP BY argument;
`

// getDataExportClickCountSql depends on the InsightsDataExportRequest ping,
// which is defined in cmd/frontend/internal/insights/httpapi/export.go
const getDataExportClickCountSql = `
SELECT COUNT(*) FROM event_logs
WHERE name = 'InsightsDataExportRequest' AND timestamp > DATE_TRUNC('week', $1::TIMESTAMP);
`

const InsightsTotalCountPingName = `INSIGHT_TOTAL_COUNTS`
const InsightsTotalCountCriticalPingName = `INSIGHT_TOTAL_COUNT_CRITICAL`
const InsightsIntervalCountsPingName = `INSIGHT_TIME_INTERVALS`
const InsightsOrgVisibleInsightsPingName = `INSIGHT_ORG_VISIBLE_INSIGHTS`
const InsightsTotalOrgsWithDashboardPingName = `INSIGHT_TOTAL_ORGS_WITH_DASHBOARD`
const InsightsDashboardTotalCountPingName = `INSIGHT_DASHBOARD_TOTAL_COUNT`
const InsightsPerDashboardPingName = `INSIGHTS_PER_DASHBORD_STATS`
const InsightsBackfillTimePingName = `INSIGHTS_BACKFILL_TIME`

package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodeInsightsUsageStatistics(ctx context.Context) (*types.CodeInsightsUsageStatistics, error) {
	stats := types.CodeInsightsUsageStatistics{}

	const viewingMetricsQuery = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'ViewInsights')                       AS insights_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsights')        AS insights_unique_page_views,
		COUNT(*) FILTER
			(WHERE name = 'InsightHover'
				AND argument ->> 'insightType'::text = 'searchInsights')    AS search_insights_hovers,
		COUNT(*) FILTER
			(WHERE name = 'InsightHover'
				AND argument ->> 'insightType'::text = 'codeStatsInsights') AS code_stats_insights_hovers,
		COUNT(*) FILTER (WHERE name = 'InsightUICustomization')             AS insights_ui_customizations,
		COUNT(*) FILTER (WHERE name = 'InsightDataPointClick')              AS insights_data_point_clicks
	FROM event_logs
	WHERE name in ('ViewInsights', 'InsightHover', 'InsightUICustomization', 'InsightDataPointClick');
	`

	if err := dbconn.Global.QueryRowContext(ctx, viewingMetricsQuery).Scan(
		&stats.InsightsPageViews,
		&stats.InsightsUniquePageViews,
		&stats.SearchInsightsHovers,
		&stats.CodeStatsInsightsHovers,
		&stats.InsightsUICustomizations,
		&stats.InsightsDataPointClicks,
	); err != nil {
		return nil, err
	}

	const creationMetricsQuery = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'InsightEdit'
			AND argument ->> 'insightType'::text = 'codeStatsInsights') AS code_stats_insights_edits,
		COUNT(*) FILTER (WHERE name = 'InsightAddition'
			AND argument ->> 'insightType'::text = 'codeStatsInsights') AS code_stats_insights_additions,
		COUNT(*) FILTER (WHERE name = 'InsightRemoval'
			AND argument ->> 'insightType'::text = 'codeStatsInsights') AS code_stats_insights_removals,

		COUNT(*) FILTER (WHERE name = 'InsightEdit'
			AND argument ->> 'insightType'::text = 'searchInsights') AS search_insights_edits,
		COUNT(*) FILTER (WHERE name = 'InsightAddition'
			AND argument ->> 'insightType'::text = 'searchInsights') AS search_insights_additions,
		COUNT(*) FILTER (WHERE name = 'InsightRemoval'
			AND argument ->> 'insightType'::text = 'searchInsights') AS search_insights_removals,

		COUNT(distinct anonymous_user_id) FILTER (WHERE name = 'InsightAddition' AND timestamp > DATE_TRUNC('week', $1::timestamp))
			AS weekly_insight_creators,

		COUNT(*) FILTER (WHERE name = 'InsightConfigureClick') AS insight_configure_click,
		COUNT(*) FILTER (WHERE name = 'InsightAddMoreClick') AS insight_add_more_click
	FROM event_logs
	WHERE name in ('InsightEdit', 'InsightAddition', 'InsightRemoval', 'InsightConfigureClick', 'InsightAddMoreClick');
	`

	if err := dbconn.Global.QueryRowContext(ctx, creationMetricsQuery, timeNow()).Scan(
		&stats.CodeStatsInsightsEdits,
		&stats.CodeStatsInsightsAdditions,
		&stats.CodeStatsInsightsRemovals,
		&stats.SearchInsightsEdits,
		&stats.SearchInsightsAdditions,
		&stats.SearchInsightsRemovals,
		&stats.WeeklyInsightCreators,
		&stats.InsightConfigureClick,
		&stats.InsightAddMoreClick,
	); err != nil {
		return nil, err
	}

	const weeklyFirstTimeCreatorsQuery = `
	WITH first_times AS (
		SELECT
			anonymous_user_id,
			MIN(timestamp) as first_time
		FROM event_logs
		WHERE name = 'InsightAddition'
		GROUP BY anonymous_user_id
		)
	SELECT
		DATE_TRUNC('week', $1::timestamp) AS week_start,
		COUNT(distinct anonymous_user_id) as weekly_first_time_insight_creators
	FROM first_times
	WHERE first_time > DATE_TRUNC('week', $1::timestamp);
	`

	if err := dbconn.Global.QueryRowContext(ctx, weeklyFirstTimeCreatorsQuery, timeNow()).Scan(
		&stats.WeekStart,
		&stats.WeeklyFirstTimeInsightCreators,
	); err != nil {
		return nil, err
	}

	return &stats, nil
}

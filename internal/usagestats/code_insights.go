package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodeInsightsUsageStatistics(ctx context.Context) (*types.CodeInsightsUsageStatistics, error) {
	stats := types.CodeInsightsUsageStatistics{}

	const platformQuery = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'ViewInsights')                       AS insights_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsights')        AS insights_unique_page_views,
		COUNT(distinct anonymous_user_id)
			FILTER (WHERE name = 'InsightAddition'
				AND timestamp > DATE_TRUNC('week', $1::timestamp))			AS weekly_insight_creators,
		COUNT(*) FILTER (WHERE name = 'InsightConfigureClick') 				AS insight_configure_click,
		COUNT(*) FILTER (WHERE name = 'InsightAddMoreClick') 				AS insight_add_more_click
	FROM event_logs
	WHERE name in ('ViewInsights', 'InsightAddition', 'InsightConfigureClick', 'InsightAddMoreClick');
	`

	if err := dbconn.Global.QueryRowContext(ctx, platformQuery, timeNow()).Scan(
		&stats.InsightsPageViews,
		&stats.InsightsUniquePageViews,
		&stats.WeeklyInsightCreators,
		&stats.InsightConfigureClick,
		&stats.InsightAddMoreClick,
	); err != nil {
		return nil, err
	}

	const metricsByInsightQuery = `
	SELECT argument ->> 'insightType'::text 					AS insight_type,
        COUNT(*) FILTER (WHERE name = 'InsightAddition') 		AS additions,
        COUNT(*) FILTER (WHERE name = 'InsightEdit') 			AS edits,
        COUNT(*) FILTER (WHERE name = 'InsightRemoval') 		AS removals,
		COUNT(*) FILTER (WHERE name = 'InsightHover') 			AS hovers,
		COUNT(*) FILTER (WHERE name = 'InsightUICustomization') AS ui_customizations,
		COUNT(*) FILTER (WHERE name = 'InsightDataPointClick') 	AS data_point_clicks
	FROM event_logs
	WHERE name in ('InsightAddition', 'InsightEdit', 'InsightRemoval', 'InsightHover', 'InsightUICustomization', 'InsightDataPointClick')
	GROUP BY insight_type;
	`

	usageStatisticsByInsight := []*types.InsightUsageStatistics{}
	rows, err := dbconn.Global.QueryContext(ctx, metricsByInsightQuery)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		insightUsageStatistics := types.InsightUsageStatistics{}

		if err := rows.Scan(
			&insightUsageStatistics.InsightType,
			&insightUsageStatistics.Additions,
			&insightUsageStatistics.Edits,
			&insightUsageStatistics.Removals,
			&insightUsageStatistics.Hovers,
			&insightUsageStatistics.UICustomizations,
			&insightUsageStatistics.DataPointClicks,
		); err != nil {
			return nil, err
		}

		usageStatisticsByInsight = append(usageStatisticsByInsight, &insightUsageStatistics)
	}
	stats.UsageStatisticsByInsight = usageStatisticsByInsight

	if err := rows.Err(); err != nil {
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

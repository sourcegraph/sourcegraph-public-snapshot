package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodeInsightsUsageStatistics(ctx context.Context, db dbutil.DB) (*types.CodeInsightsUsageStatistics, error) {
	stats := types.CodeInsightsUsageStatistics{}

	const platformQuery = `
	SELECT
		COUNT(*) FILTER (WHERE name = 'ViewInsights')                       AS weekly_insights_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsights')        AS weekly_insights_unique_page_views,
		COUNT(distinct anonymous_user_id)
			FILTER (WHERE name = 'InsightAddition')							AS weekly_insight_creators,
		COUNT(*) FILTER (WHERE name = 'InsightConfigureClick') 				AS weekly_insight_configure_click,
		COUNT(*) FILTER (WHERE name = 'InsightAddMoreClick') 				AS weekly_insight_add_more_click
	FROM event_logs
	WHERE name in ('ViewInsights', 'InsightAddition', 'InsightConfigureClick', 'InsightAddMoreClick')
		AND timestamp > DATE_TRUNC('week', $1::timestamp);
	`

	if err := db.QueryRowContext(ctx, platformQuery, timeNow()).Scan(
		&stats.WeeklyInsightsPageViews,
		&stats.WeeklyInsightsUniquePageViews,
		&stats.WeeklyInsightCreators,
		&stats.WeeklyInsightConfigureClick,
		&stats.WeeklyInsightAddMoreClick,
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

	if err := db.QueryRowContext(ctx, weeklyFirstTimeCreatorsQuery, timeNow()).Scan(
		&stats.WeekStart,
		&stats.WeeklyFirstTimeInsightCreators,
	); err != nil {
		return nil, err
	}

	return &stats, nil
}

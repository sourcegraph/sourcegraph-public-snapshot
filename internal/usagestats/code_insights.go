package usagestats

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"

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

	weeklyUsage, err := GetCreationViewUsage(ctx, db, timeNow)
	if err != nil {
		return nil, err
	}
	stats.WeeklyAggregatedUsage = weeklyUsage

	timeIntervals, err := GetInsightTimeIntervals(ctx, db, timeNow)
	if err != nil {
		return nil, err
	}
	stats.InsightTimeIntervals = timeIntervals

	orgVisible, err := GetInsightCountsByOrg(ctx, db, timeNow)
	if err != nil {
		return nil, err
	}
	stats.InsightOrgVisible = orgVisible

	return &stats, nil
}

func GetCreationViewUsage(ctx context.Context, db dbutil.DB, timeSupplier func() time.Time) ([]types.AggregatedPingStats, error) {
	builder := creationPagesPingBuilder(timeSupplier)

	results, err := builder.Sample(ctx, db)
	if err != nil {
		return []types.AggregatedPingStats{}, err
	}

	return results, nil
}

func GetInsightTimeIntervals(ctx context.Context, db dbutil.DB, timeSupplier func() time.Time) ([]types.InsightTimeIntervalPing, error) {
	//	query for
	rows, err := db.QueryContext(ctx, insightTimeIntervalQueryStr, timeSupplier())
	if err != nil {
		return []types.InsightTimeIntervalPing{}, err
	}
	defer rows.Close()

	results := make([]types.InsightTimeIntervalPing, 0)

	for rows.Next() {
		var temp types.InsightTimeIntervalPing
		if err := rows.Scan(&temp.IntervalDays, &temp.TotalCount); err != nil {
			return []types.InsightTimeIntervalPing{}, err
		}
		results = append(results, temp)
	}

	return results, nil
}

func GetInsightCountsByOrg(ctx context.Context, db dbutil.DB, timeSupplier func() time.Time) ([]types.OrgVisibleInsightPing, error) {
	//	query for
	rows, err := db.QueryContext(ctx, insightOrgVisiblePingQueryStr, timeSupplier())
	if err != nil {
		return []types.OrgVisibleInsightPing{}, err
	}
	defer rows.Close()

	results := make([]types.OrgVisibleInsightPing, 0)

	for rows.Next() {
		var temp types.OrgVisibleInsightPing
		if err := rows.Scan(&temp.Type, &temp.TotalCount); err != nil {
			return []types.OrgVisibleInsightPing{}, err
		}
		results = append(results, temp)
	}

	return results, nil
}

// WithAll add multiple pings by name to this builder
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

// Sample execute the derived query generated by this builder and return a sample at the current time
func (b *PingQueryBuilder) Sample(ctx context.Context, db dbutil.DB) ([]types.AggregatedPingStats, error) {

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
	Day              = "day"
	Week             = "week"
	Month            = "month"
	Year             = "year"
)

const templatePingQueryStr = `
select name, count(*) as total_count, count(distinct anonymous_user_id) as unique_count
from event_logs
where name = any($2)
  AND timestamp > DATE_TRUNC('%v', $1::timestamp)
group by name;
`

const insightTimeIntervalQueryStr = `
select JSON_ARRAY_ELEMENTS(argument::json)::text as interval_days, count(*) from event_logs
join (select max(id) as id from event_logs where name = 'InsightsGroupedStepSizes') as most_recent_event
on most_recent_event.id = event_logs.id
where name = 'InsightsGroupedStepSizes'
and timestamp > DATE_TRUNC('week', $1::timestamp)
group by name, interval_days;
`

const insightOrgVisiblePingQueryStr = `
select flattened.key as type, flattened.value as total_count from event_logs
join json_each_text(event_logs.argument::json) as flattened on true
join (select max(id) as id from event_logs where name = 'InsightsGroupedCount') as most_recent_event
     on most_recent_event.id = event_logs.id
where event_logs.name = 'InsightsGroupedCount'
and timestamp > DATE_TRUNC('week', $1::timestamp);`

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
		COUNT(*) FILTER (WHERE name = 'ViewInsights')                       AS weekly_insights_page_views,
		COUNT(distinct user_id) FILTER (WHERE name = 'ViewInsights')        AS weekly_insights_unique_page_views,
		COUNT(distinct user_id)
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

	return &stats, nil
}

func GetTotalInsightCounts(ctx context.Context, db database.DB) (types.InsightTotalCounts, error) {
	store := database.EventLogs(db)
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
	store := database.EventLogs(db)
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
	store := database.EventLogs(db)
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

func GetCreationViewUsage(ctx context.Context, db database.DB, timeSupplier func() time.Time) ([]types.AggregatedPingStats, error) {
	builder := creationPagesPingBuilder(timeSupplier)

	results, err := builder.Sample(ctx, db)
	if err != nil {
		return []types.AggregatedPingStats{}, err
	}

	return results, nil
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

const InsightsTotalCountPingName = `INSIGHT_TOTAL_COUNTS`
const InsightsIntervalCountsPingName = `INSIGHT_TIME_INTERVALS`
const InsightsOrgVisibleInsightsPingName = `INSIGHT_ORG_VISIBLE_INSIGHTS`

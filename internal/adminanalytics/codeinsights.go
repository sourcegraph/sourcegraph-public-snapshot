package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type CodeInsights struct {
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     bool
}

var insightsCreatedNodesQuery = `
	SELECT
		%s AS date,
		COUNT(DISTINCT s.id) AS count,
		COUNT(DISTINCT g.user_id) AS unique_users,
		COUNT(DISTINCT g.user_id) AS registered_users
	FROM insight_series AS s
		JOIN insight_view_series AS vs ON vs.insight_series_id = s.id
		LEFT JOIN insight_view_grants AS g ON g.insight_view_id = vs.insight_view_id
	WHERE %s
	GROUP BY date
`

var insightsCreatedSummaryQuery = `
	SELECT
		COUNT(DISTINCT s.id) AS total_count,
		COUNT(DISTINCT g.user_id) AS total_unique_users,
		COUNT(DISTINCT g.user_id) AS total_registered_users
	FROM insight_series AS s
		JOIN insight_view_series AS vs ON vs.insight_series_id = s.id
		LEFT JOIN insight_view_grants AS g ON g.insight_view_id = vs.insight_view_id
	WHERE %s
`

func (c *CodeInsights) InsightCreations() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(c.DateRange, c.Grouping, "s.created_at")
	if err != nil {
		return nil, err
	}
	conds := []*sqlf.Query{sqlf.Sprintf(`s.created_at %s`, dateBetweenCond)}

	nodesQuery := sqlf.Sprintf(insightsCreatedNodesQuery, dateTruncExp, sqlf.Join(conds, "AND"))
	summaryQuery := sqlf.Sprintf(insightsCreatedSummaryQuery, sqlf.Join(conds, "AND"))

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Insights:InsightCreations",
		cache:        c.Cache,
	}, nil
}

func (c *CodeInsights) InsightHovers() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		c.DateRange,
		c.Grouping,
		[]string{"InsightHover"},
	)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Insights:InsightsHovers",
		cache:        c.Cache,
	}, nil
}

func (c *CodeInsights) InsightDataPointClicks() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		c.DateRange,
		c.Grouping,
		[]string{"InsightDataPointClick"},
	)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Insights:InsightsDataPointClicks",
		cache:        c.Cache,
	}, nil
}

func (c *CodeInsights) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){c.InsightHovers, c.InsightDataPointClicks}
	for _, buildFetcher := range fetcherBuilders {
		fetcher, err := buildFetcher()
		if err != nil {
			return err
		}

		if _, err := fetcher.Nodes(ctx); err != nil {
			return err
		}

		if _, err := fetcher.Summary(ctx); err != nil {
			return err
		}
	}
	return nil
}

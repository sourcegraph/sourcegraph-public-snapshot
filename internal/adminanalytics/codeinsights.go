package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CodeInsights struct {
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     bool
}

// Insights:SeriesCreations

var seriesCreationNodesQuery = `
	SELECT
		%s AS date,
		COUNT(DISTINCT id) AS count,
		-- Add empty columns to reuse AnalyticsFetcher
		0 as unique_users,
		0 as registered_users
	FROM insight_series
	WHERE %s
	GROUP BY date
`

var seriesCreationsSummaryQuery = `
	SELECT
		COUNT(DISTINCT id) AS total_count,
		-- Add empty columns to reuse AnalyticsFetcher
		0 as total_unique_users,
		0 as total_registered_users
	FROM insight_series
	WHERE %s
`

func makeGenerationTypeField(generationType string) (string, error) {
	switch generationType {
	case "SEARCH":
		return "search", nil
	case "SEARCH_COMPUTE":
		return "search-compute", nil
	case "LANGUAGE_STATS":
		return "language-stats", nil
	default:
		return "", errors.Newf("Unknown code insights generation type: %s", generationType)
	}
}

func (c *CodeInsights) SeriesCreations(ctx context.Context, args *struct {
	GenerationType *string
}) (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(c.DateRange, c.Grouping, "created_at")
	if err != nil {
		return nil, err
	}
	conds := []*sqlf.Query{sqlf.Sprintf(`created_at %s`, dateBetweenCond)}
	if args.GenerationType != nil {
		generationType, err := makeGenerationTypeField(*args.GenerationType)
		if err != nil {
			return nil, err
		}
		conds = append(conds, sqlf.Sprintf(`generation_method = %s`, generationType))
	}

	nodesQuery := sqlf.Sprintf(seriesCreationNodesQuery, dateTruncExp, sqlf.Join(conds, "AND"))
	summaryQuery := sqlf.Sprintf(seriesCreationsSummaryQuery, sqlf.Join(conds, "AND"))

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Insights:SeriesCreations",
		cache:        c.Cache,
	}, nil
}

// Insights:DashboardCreations

var dashboardCreationNodesQuery = `
	SELECT
		%s AS date,
		COUNT(DISTINCT id) AS count,
		-- Add empty columns to reuse AnalyticsFetcher
		0 as unique_users,
		0 as registered_users
	FROM dashboard
	WHERE %s
	GROUP BY date
`

var dashboardCreationsSummaryQuery = `
	SELECT
		COUNT(DISTINCT id) AS total_count,
		-- Add empty columns to reuse AnalyticsFetcher
		0 as total_unique_users,
		0 as total_registered_users
	FROM dashboard
	WHERE %s
`

func (c *CodeInsights) DashboardCreations() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(c.DateRange, c.Grouping, "created_at")
	if err != nil {
		return nil, err
	}
	conds := []*sqlf.Query{sqlf.Sprintf(`created_at %s`, dateBetweenCond)}
	nodesQuery := sqlf.Sprintf(dashboardCreationNodesQuery, dateTruncExp, sqlf.Join(conds, "AND"))
	summaryQuery := sqlf.Sprintf(dashboardCreationsSummaryQuery, sqlf.Join(conds, "AND"))

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Insights:DashboardCreations",
		cache:        c.Cache,
	}, nil
}

// Insights:InsightCreations

var insightCTEQueryFragment = `
WITH insights AS (
    SELECT
		v.id as insight_id,
		MIN(s.created_at) as created_at
	FROM insight_view v
        INNER JOIN insight_view_series vs ON vs.insight_view_id = v.id
        INNER JOIN insight_series s ON s.id = vs.insight_series_id
    GROUP BY v.id
)
`

var insightCreationNodesQuery = insightCTEQueryFragment + `
	SELECT
		%s AS date,
		COUNT(DISTINCT insight_id) AS count,
		-- Add empty columns to reuse AnalyticsFetcher
		0 as unique_users,
		0 as registered_users
		FROM insights
	WHERE %s
	GROUP BY date
`

var insightCreationsSummaryQuery = insightCTEQueryFragment + `
	SELECT
		COUNT(DISTINCT insight_id) AS count,
		-- Add empty columns to reuse AnalyticsFetcher
		0 as unique_users,
		0 as registered_users
	FROM insights
	WHERE %s
`

func (c *CodeInsights) InsightCreations() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(c.DateRange, c.Grouping, "created_at")
	if err != nil {
		return nil, err
	}
	conds := []*sqlf.Query{sqlf.Sprintf(`created_at %s`, dateBetweenCond)}
	nodesQuery := sqlf.Sprintf(insightCreationNodesQuery, dateTruncExp, sqlf.Join(conds, "AND"))
	summaryQuery := sqlf.Sprintf(insightCreationsSummaryQuery, sqlf.Join(conds, "AND"))

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

// Insights:Hovers

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
		group:        "Insights:InsightHovers",
		cache:        c.Cache,
	}, nil
}

// Insights:DataPointClicks

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
		group:        "Insights:InsightDataPointClicks",
		cache:        c.Cache,
	}, nil
}

// Insights caching job entrypoint

func (c *CodeInsights) CacheAll(ctx context.Context) error {
	// TODO: make sure to add all entry points
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

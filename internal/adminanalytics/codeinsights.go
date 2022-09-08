package adminanalytics

import (
	"context"
	"fmt"

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

var (
	searchGenerationType        = "SEARCH"
	searchComputeGenerationType = "SEARCH_COMPUTE"
	languageStatsGenerationType = "LANGUAGE_STATS"
)

func makeGenerationTypeField(generationType string) (string, error) {
	switch generationType {
	case searchGenerationType:
		return "search", nil
	case searchComputeGenerationType:
		return "search-compute", nil
	case languageStatsGenerationType:
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
	cacheGroupKey := "Insights:SeriesCreations"
	if args.GenerationType != nil {
		generationType, err := makeGenerationTypeField(*args.GenerationType)
		if err != nil {
			return nil, err
		}
		conds = append(conds, sqlf.Sprintf(`generation_method = %s`, generationType))
		cacheGroupKey = fmt.Sprintf("%s:%s", cacheGroupKey, *args.GenerationType)
	} else {
		cacheGroupKey = fmt.Sprintf("%s:%s", cacheGroupKey, "ALL")
	}

	nodesQuery := sqlf.Sprintf(seriesCreationNodesQuery, dateTruncExp, sqlf.Join(conds, "AND"))
	summaryQuery := sqlf.Sprintf(seriesCreationsSummaryQuery, sqlf.Join(conds, "AND"))

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        cacheGroupKey,
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

// Insights:TotalInsightsCount

func (c *CodeInsights) TotalInsightsCount(ctx context.Context) (*float64, error) {
	cacheKey := "Insights:TotalInsightsCount"
	if c.Cache {
		if totalCount, err := getItemFromCache[float64](cacheKey); err == nil {
			return totalCount, nil
		}
	}

	var totalCount float64
	query := `SELECT COUNT (distinct id) AS total_count FROM insight_view`
	if err := c.DB.QueryRowContext(ctx, query).Scan(&totalCount); err != nil {
		return nil, err
	}

	if _, err := setItemToCache(cacheKey, &totalCount); err != nil {
		return nil, err
	}

	return &totalCount, nil
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
	// Cache general insights stats based on AnalyticsFetcher
	fetcherBuilders := []func() (*AnalyticsFetcher, error){c.DashboardCreations, c.InsightHovers, c.InsightDataPointClicks}
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

	// Cache total insights count
	if _, err := c.TotalInsightsCount(ctx); err != nil {
		return err
	}

	// Cache series creation stats
	for _, generationType := range []*string{nil, &searchComputeGenerationType, &searchGenerationType, &languageStatsGenerationType} {
		fetcher, err := c.SeriesCreations(ctx, &struct{ GenerationType *string }{GenerationType: generationType})
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

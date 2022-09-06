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
		COUNT(DISTINCT changesets.id) AS count,
		COUNT(DISTINCT batch_changes.creator_id) AS unique_users,
		COUNT(DISTINCT batch_changes.creator_id) AS registered_users
	FROM
		changesets
		INNER JOIN batch_changes ON batch_changes.id = changesets.owned_by_batch_change_id
	WHERE changesets.created_at %s AND changesets.publication_state = 'PUBLISHED'
	GROUP BY date
`

var insightsCreatedSummaryQuery = `
	SELECT
		COUNT(DISTINCT changesets.id) AS total_count,
		COUNT(DISTINCT batch_changes.creator_id) AS total_unique_users,
		COUNT(DISTINCT batch_changes.creator_id) AS total_registered_users
	FROM
		changesets
		INNER JOIN batch_changes ON batch_changes.id = changesets.owned_by_batch_change_id
	WHERE changesets.created_at %s AND changesets.publication_state = 'PUBLISHED'
`

func (c *CodeInsights) InsightCreated() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(c.DateRange, c.Grouping, "changesets.created_at")
	if err != nil {
		return nil, err
	}

	nodesQuery := sqlf.Sprintf(insightsCreatedNodesQuery, dateTruncExp, dateBetweenCond)
	summaryQuery := sqlf.Sprintf(insightsCreatedSummaryQuery, dateBetweenCond)

	return &AnalyticsFetcher{
		db:           c.DB,
		dateRange:    c.DateRange,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Insights:InsightCreated",
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

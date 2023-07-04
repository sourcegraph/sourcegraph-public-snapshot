package adminanalytics

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Cody struct {
	Ctx       context.Context
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     bool
}

// These conditions match those found in the aggregatedCodyUsageEventsQuery variable
// in /internal/database/event_logs.go. Ensure these two filters remain in sync to ensure
// that Sourcegraph and customers see the same data.
var codySourceCond = `lower(name) like '%%cody%%'
	AND name not like '%%CTA%%'
	AND name not like '%%Cta%%'
	AND (name NOT IN ('` + strings.Join(database.NonActiveCodyEvents, "','") + `'))`

// Cody:Users

func (c *Cody) Users() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		c.DateRange,
		c.Grouping,
		[]string{},
		sqlf.Sprintf(codySourceCond),
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
		group:        "Cody:Users",
		cache:        c.Cache,
	}, nil
}

// Cody:Prompts

func (c *Cody) Prompts() (*AnalyticsFetcher, error) {
	sourceCond := sqlf.Sprintf(codySourceCond + ` AND lower(name) like '%%recipe%%'`)

	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		c.DateRange,
		c.Grouping,
		[]string{},
		sourceCond,
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
		group:        "Cody:Prompts",
		cache:        c.Cache,
	}, nil
}

// Cody:CompletionsSuggested

func (c *Cody) CompletionsSuggested() (*AnalyticsFetcher, error) {
	sourceCond := sqlf.Sprintf(`CAST(argument->>'displayDuration' AS INTEGER) >= 750`)

	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		c.DateRange,
		c.Grouping,
		[]string{"CodyVSCodeExtension:completion:suggested"},
		sourceCond,
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
		group:        "Cody:CompletionsSuggested",
		cache:        c.Cache,
	}, nil
}

// Cody:CompletionsAccepted

func (c *Cody) CompletionsAccepted() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		c.DateRange,
		c.Grouping,
		[]string{"CodyVSCodeExtension:completion:accepted"},
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
		group:        "Cody:CompletionsAccepted",
		cache:        c.Cache,
	}, nil
}

// Cody caching job entrypoint

func (c *Cody) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){c.Users, c.Prompts, c.CompletionsAccepted, c.CompletionsSuggested}
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

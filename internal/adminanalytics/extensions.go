package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Extensions struct {
	Ctx       context.Context
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     KeyValue
}

func (e *Extensions) Jetbrains() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		e.DateRange,
		e.Grouping,
		[]string{"IDESearchSubmitted", "VSCESearchSubmitted"},
		sqlf.Sprintf("source = 'IDEEXTENSION' AND referrer = 'JETBRAINS'"),
	)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           e.DB,
		dateRange:    e.DateRange,
		grouping:     e.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Extensions:Jetbrains",
		cache:        NoopCache{},
	}, nil
}

func (e *Extensions) Vscode() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		e.DateRange,
		e.Grouping,
		[]string{"IDESearchSubmitted", "VSCESearchSubmitted"},
		sqlf.Sprintf("source = 'IDEEXTENSION' AND referrer = 'VSCE'"),
	)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           e.DB,
		dateRange:    e.DateRange,
		grouping:     e.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Extensions:Vscode",
		cache:        e.Cache,
	}, nil
}

func (e *Extensions) Browser() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(
		e.DateRange,
		e.Grouping,
		[]string{"goToDefinition.preloaded", "goToDefinition", "findReferences"},
		sqlf.Sprintf("source = 'CODEHOSTINTEGRATION'"),
	)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           e.DB,
		dateRange:    e.DateRange,
		grouping:     e.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Extensions:Browser",
		cache:        e.Cache,
	}, nil
}

func (e *Extensions) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){e.Jetbrains, e.Vscode, e.Browser}
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

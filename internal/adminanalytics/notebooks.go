package adminanalytics

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Notebooks struct {
	DateRange string
	DB        database.DB
}

func (s *Notebooks) Creations() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"SearchNotebookCreated"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Notebooks:Creations",
	}, nil
}

func (s *Notebooks) Views() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"SearchNotebookPageViewed"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Notebooks:Views",
	}, nil
}

func (s *Notebooks) BlockRuns() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{
		"SearchNotebookRunAllBlocks",
		"SearchNotebookRunBlock",
	})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Notebooks:BlockRuns",
	}, nil
}

func (s *Notebooks) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){s.Creations, s.BlockRuns, s.Views}
	for _, buildFetcher := range fetcherBuilders {
		fetcher, err := buildFetcher()
		if err != nil {
			return err
		}

		if _, err := fetcher.GetNodes(ctx, false); err != nil {
			return err
		}

		if _, err := fetcher.GetSummary(ctx, false); err != nil {
			return err
		}
	}
	return nil
}

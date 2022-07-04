package adminanalytics

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type CodeIntel struct {
	DateRange string
	DB        database.DB
}

func (s *CodeIntel) ReferenceClicks() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"findReferences"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:ReferenceClicks",
	}, nil
}

func (s *CodeIntel) DefinitionClicks() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"goToDefinition.preloaded", "goToDefinition"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:DefinitionClicks",
	}, nil
}

func (s *CodeIntel) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){s.DefinitionClicks, s.ReferenceClicks}
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

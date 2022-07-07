package adminanalytics

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type CodeIntel struct {
	DateRange string
	DB        database.DB
	Cache     bool
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
		cache:        s.Cache,
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
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) BrowserExtensionInstalls() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"BrowserExtensionInstalled"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:BrowserExtensionInstalls",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) SearchBasedEvents() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{
		"codeintel.searchDefinitions",
		"codeintel.searchDefinitions.xrepo",
		"codeintel.searchReferences",
		"codeintel.searchReferences.xrepo",
	})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:SearchBasedEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) PreciseEvents() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{
		"codeintel.lsifDefinitions",
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences",
		"codeintel.lsifReferences.xrepo",
	})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:PreciseEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) CrossRepoEvents() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{
		"codeintel.searchDefinitions.xrepo",
		"codeintel.searchReferences.xrepo",
		"codeintel.lsifDefinitions.xrepo",
		"codeintel.lsifReferences.xrepo",
	})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:CrossRepoEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){
		s.DefinitionClicks,
		s.ReferenceClicks,
		s.BrowserExtensionInstalls,
		s.SearchBasedEvents,
		s.PreciseEvents,
		s.CrossRepoEvents,
	}

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

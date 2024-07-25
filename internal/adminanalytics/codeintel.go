package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type CodeIntel struct {
	Ctx       context.Context
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     KeyValue
}

func (s *CodeIntel) ReferenceClicks() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{"findReferences"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:ReferenceClicks",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) DefinitionClicks() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{"goToDefinition.preloaded", "goToDefinition"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:DefinitionClicks",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) InAppEvents() (*AnalyticsFetcher, error) {
	sourceCond := sqlf.Sprintf("source = 'WEB'")
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{"goToDefinition.preloaded", "goToDefinition", "findReferences"}, sourceCond)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:InAppEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) CodeHostEvents() (*AnalyticsFetcher, error) {
	sourceCond := sqlf.Sprintf("source = 'CODEHOSTINTEGRATION'")
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{"goToDefinition.preloaded", "goToDefinition", "findReferences"}, sourceCond)
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:CodeHostEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) SearchBasedEvents() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{
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
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:SearchBasedEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) PreciseEvents() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{
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
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "CodeIntel:PreciseEvents",
		cache:        s.Cache,
	}, nil
}

func (s *CodeIntel) CrossRepoEvents() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{
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
		grouping:     s.Grouping,
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
		s.InAppEvents,
		s.CodeHostEvents,
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

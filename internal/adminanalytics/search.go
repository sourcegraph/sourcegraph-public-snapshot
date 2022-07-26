package adminanalytics

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Search struct {
	DateRange string
	DB        database.DB
	Cache     bool
}

func (s *Search) Searches() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"SearchResultsQueried"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:Searches",
		cache:        s.Cache,
	}, nil
}

func (s *Search) ResultClicks() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"SearchResultClicked"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:ResultClicked",
	}, nil
}

func (s *Search) FileViews() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{"ViewBlob"})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:FileViews",
		cache:        s.Cache,
	}, nil
}

func (s *Search) FileOpens() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{
		"GoToCodeHostClicked",
		"vscode.open.file",
		"openInAtom.open.file",
		"openineditor.open.file",
		"openInWebstorm.open.file",
	})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:FileOpens",
		cache:        s.Cache,
	}, nil
}

func (s *Search) CacheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnalyticsFetcher, error){s.Searches, s.FileViews, s.FileOpens, s.ResultClicks}
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

type csvEvent struct {
	Name    string
	Fetcher func() (*AnalyticsFetcher, error)
}

func (s *Search) ExportCSV(ctx context.Context) (*string, error) {
	rows := []string{"Event,Metric,Date,Value"}

	events := []csvEvent{
		{Name: "Searches", Fetcher: s.Searches},
		{Name: "Result Clicks", Fetcher: s.ResultClicks},
		{Name: "File Views", Fetcher: s.FileViews},
	}

	for _, event := range events {
		if fetcher, err := event.Fetcher(); err == nil {
			if nodes, err := fetcher.Nodes(ctx); err == nil {
				for _, node := range nodes {
					rows = append(rows, fmt.Sprintf("%s,Total Events,%s,%v", event.Name, node.Date(), node.Count()))
					rows = append(rows, fmt.Sprintf("%s,Unique Users,%s,%v", event.Name, node.Date(), node.UniqueUsers()))
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}

	}

	csv := strings.Join(rows[:], "\n")

	return &csv, nil
}

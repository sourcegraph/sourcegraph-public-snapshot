package adminanalytics

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type AnalyticsFetcher struct {
	db           database.DB
	group        string
	dateRange    string
	grouping     string
	nodesQuery   *sqlf.Query
	summaryQuery *sqlf.Query
	cache        KeyValue
}

type AnalyticsNodeData struct {
	Date            time.Time
	Count           float64
	UniqueUsers     float64
	RegisteredUsers float64
}

type AnalyticsNode struct {
	Data AnalyticsNodeData
}

func (n *AnalyticsNode) Date() string { return n.Data.Date.Format(time.RFC3339) }

func (n *AnalyticsNode) Count() float64 { return n.Data.Count }

func (n *AnalyticsNode) UniqueUsers() float64 { return n.Data.UniqueUsers }

func (n *AnalyticsNode) RegisteredUsers() float64 { return n.Data.RegisteredUsers }

func (f *AnalyticsFetcher) Nodes(ctx context.Context) ([]*AnalyticsNode, error) {
	cacheKey := fmt.Sprintf(`%s:%s:%s:%s`, f.group, f.dateRange, f.grouping, "nodes")

	if nodes, err := getArrayFromCache[AnalyticsNode](f.cache, cacheKey); err == nil {
		return nodes, nil
	}

	rows, err := f.db.QueryContext(ctx, f.nodesQuery.Query(sqlf.PostgresBindVar), f.nodesQuery.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*AnalyticsNode, 0)
	for rows.Next() {
		var data AnalyticsNodeData

		if err := rows.Scan(&data.Date, &data.Count, &data.UniqueUsers, &data.RegisteredUsers); err != nil {
			return nil, err
		}

		nodes = append(nodes, &AnalyticsNode{data})
	}

	now := time.Now()
	to := now
	daysOffset := 1
	from, err := getFromDate(f.dateRange, now)
	if err != nil {
		return nil, err
	}

	if f.grouping == Weekly {
		to = now.AddDate(0, 0, -int(now.Weekday())+1) // monday of current week
		daysOffset = 7
	}

	allNodes := make([]*AnalyticsNode, 0)

	for date := to; date.After(from) || date.Equal(from); date = date.AddDate(0, 0, -daysOffset) {
		var node *AnalyticsNode

		for _, n := range nodes {
			if bod(date).Equal(bod(n.Data.Date)) {
				node = n
				break
			}
		}

		if node == nil {
			node = &AnalyticsNode{
				Data: AnalyticsNodeData{
					Date:            bod(date),
					Count:           0,
					UniqueUsers:     0,
					RegisteredUsers: 0,
				},
			}
		}

		allNodes = append(allNodes, node)
	}

	err = setArrayToCache(f.cache, cacheKey, allNodes)
	if err != nil {
		return nil, err
	}

	return allNodes, nil

}

func bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

type AnalyticsSummaryData struct {
	TotalCount           float64
	TotalUniqueUsers     float64
	TotalRegisteredUsers float64
}

type AnalyticsSummary struct {
	Data AnalyticsSummaryData
}

func (s *AnalyticsSummary) TotalCount() float64 { return s.Data.TotalCount }

func (s *AnalyticsSummary) TotalUniqueUsers() float64 { return s.Data.TotalUniqueUsers }

func (s *AnalyticsSummary) TotalRegisteredUsers() float64 { return s.Data.TotalRegisteredUsers }

func (f *AnalyticsFetcher) Summary(ctx context.Context) (*AnalyticsSummary, error) {
	cacheKey := fmt.Sprintf(`%s:%s:%s:%s`, f.group, f.dateRange, f.grouping, "summary")
	if summary, err := getItemFromCache[AnalyticsSummary](f.cache, cacheKey); err == nil {
		return summary, nil
	}

	var data AnalyticsSummaryData

	if err := f.db.QueryRowContext(ctx, f.summaryQuery.Query(sqlf.PostgresBindVar), f.summaryQuery.Args()...).Scan(&data.TotalCount, &data.TotalUniqueUsers, &data.TotalRegisteredUsers); err != nil {
		return nil, err
	}

	summary := &AnalyticsSummary{data}

	err := setItemToCache(f.cache, cacheKey, summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

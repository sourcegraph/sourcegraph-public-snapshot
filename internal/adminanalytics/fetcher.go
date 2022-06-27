package adminanalytics

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type AnalyticsFetcher struct {
	db           database.DB
	group        string
	dateRange    string
	nodesQuery   *sqlf.Query
	summaryQuery *sqlf.Query
}

type AnalyticsNodeData struct {
	Date            time.Time
	Count           int32
	UniqueUsers     int32
	RegisteredUsers int32
}

type AnalyticsNode struct {
	Data AnalyticsNodeData
}

func (n *AnalyticsNode) Date() string {
	return n.Data.Date.Format(time.RFC3339)
}

func (n *AnalyticsNode) Count() int32 {
	return n.Data.Count
}

func (n *AnalyticsNode) UniqueUsers() int32 {
	return n.Data.UniqueUsers
}

func (n *AnalyticsNode) RegisteredUsers() int32 {
	return n.Data.RegisteredUsers
}

func (f *AnalyticsFetcher) GetNodes(ctx context.Context, cache bool) ([]*AnalyticsNode, error) {
	if cache == true {
		if nodes, err := getNodesFromCache(f); err == nil {
			return nodes, nil
		}
	}

	rows, err := f.db.QueryContext(ctx, f.nodesQuery.Query(sqlf.PostgresBindVar), f.nodesQuery.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*AnalyticsNode, 0)
	for rows.Next() {
		var date time.Time
		var count, uniqueUsers, registeredUsers int32

		if err := rows.Scan(&date, &count, &uniqueUsers, &registeredUsers); err != nil {
			return nil, err
		}

		nodes = append(nodes, &AnalyticsNode{
			AnalyticsNodeData{
				Date:            date,
				Count:           count,
				UniqueUsers:     uniqueUsers,
				RegisteredUsers: registeredUsers,
			}})
	}

	if _, err := setNodesToCache(f, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

type AnalyticsSummaryData struct {
	TotalCount           int32
	TotalUniqueUsers     int32
	TotalRegisteredUsers int32
}

type AnalyticsSummary struct {
	Data AnalyticsSummaryData
}

func (s *AnalyticsSummary) TotalCount() (int32, error) {
	return s.Data.TotalCount, nil
}

func (s *AnalyticsSummary) TotalUniqueUsers() (int32, error) {
	return s.Data.TotalUniqueUsers, nil
}

func (s *AnalyticsSummary) TotalRegisteredUsers() (int32, error) {
	return s.Data.TotalRegisteredUsers, nil
}

func (f *AnalyticsFetcher) GetSummary(ctx context.Context, cache bool) (*AnalyticsSummary, error) {
	if cache == true {
		if summary, err := getSummaryFromCache(f); err == nil {
			return summary, nil
		}
	}

	var totalCount, totalUniqueUsers, totalRegisteredUsers int32

	if err := f.db.QueryRowContext(ctx, f.summaryQuery.Query(sqlf.PostgresBindVar), f.summaryQuery.Args()...).Scan(&totalCount, &totalUniqueUsers, &totalRegisteredUsers); err != nil {
		return nil, err
	}

	summary := &AnalyticsSummary{
		AnalyticsSummaryData{
			TotalCount:           totalCount,
			TotalUniqueUsers:     totalUniqueUsers,
			TotalRegisteredUsers: totalRegisteredUsers,
		}}

	if _, err := setSummaryToCache(f, summary); err != nil {
		return nil, err
	}

	return summary, nil
}

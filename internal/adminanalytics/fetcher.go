package adminanalytics

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type AnalyticsFetcher struct {
	db           database.DB
	nodesQuery   *sqlf.Query
	summaryQuery *sqlf.Query
}

type AnalyticsNode struct {
	date            time.Time
	count           int32
	uniqueUsers     int32
	registeredUsers int32
}

func (n *AnalyticsNode) Date() string {
	return n.date.Format(time.RFC3339)
}

func (n *AnalyticsNode) Count() int32 {
	return n.count
}

func (n *AnalyticsNode) UniqueUsers() int32 {
	return n.uniqueUsers
}

func (n *AnalyticsNode) RegisteredUsers() int32 {
	return n.registeredUsers
}

func (f *AnalyticsFetcher) GetNodes(ctx context.Context) ([]*AnalyticsNode, error) {
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
			date:            date,
			count:           count,
			uniqueUsers:     uniqueUsers,
			registeredUsers: registeredUsers,
		})
	}

	return nodes, nil
}

type AnalyticsSummary struct {
	totalCount           int32
	totalUniqueUsers     int32
	totalRegisteredUsers int32
}

func (s *AnalyticsSummary) TotalCount() (int32, error) {
	return s.totalCount, nil
}

func (s *AnalyticsSummary) TotalUniqueUsers() (int32, error) {
	return s.totalUniqueUsers, nil
}

func (s *AnalyticsSummary) TotalRegisteredUsers() (int32, error) {
	return s.totalRegisteredUsers, nil
}

func (f *AnalyticsFetcher) GetSummary(ctx context.Context) (*AnalyticsSummary, error) {
	var totalCount, totalUniqueUsers, totalRegisteredUsers int32

	if err := f.db.QueryRowContext(ctx, f.summaryQuery.Query(sqlf.PostgresBindVar), f.summaryQuery.Args()...).Scan(&totalCount, &totalUniqueUsers, &totalRegisteredUsers); err != nil {
		return nil, err
	}

	return &AnalyticsSummary{
		totalCount:           totalCount,
		totalUniqueUsers:     totalUniqueUsers,
		totalRegisteredUsers: totalRegisteredUsers,
	}, nil
}

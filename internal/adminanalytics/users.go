package adminanalytics

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Users struct {
	DateRange string
	DB        database.DB
}

var (
	activitySummaryQuery = `
	SELECT
		COUNT(event_logs.*) AS total_count,
		COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
		COUNT(DISTINCT users.id) AS registered_users
	FROM users
		RIGHT JOIN event_logs ON users.id = event_logs.user_id
	WHERE event_logs.timestamp %s
	`
	activityNodesQuery = `
	SELECT %s AS date,
		COUNT(DISTINCT event_logs.*) AS total_count,
		COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
		COUNT(DISTINCT users.id) AS registered_users
	FROM users
		RIGHT JOIN event_logs ON users.id = event_logs.user_id
	WHERE event_logs.timestamp %s
	GROUP BY date
	`
)

func (s *Users) Activity() (*AnalyticsFetcher, error) {
	dateSelectParam, dateRangeCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}

	nodesQuery := sqlf.Sprintf(activityNodesQuery, dateSelectParam, dateRangeCond)

	summaryQuery := sqlf.Sprintf(activitySummaryQuery, dateRangeCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Users:Activity",
	}, nil
}

var (
	frequencyQuery = `
	WITH t1 AS (
		SELECT DATE(timestamp) AS date, anonymous_user_id AS user_id
		FROM event_logs
		WHERE DATE(timestamp) %s
		GROUP BY 2, 1
	),
	t2 AS (
		SELECT DISTINCT user_id AS user_id, COUNT(user_id) AS days_used
		FROM t1
		GROUP BY 1
	),
	t3 AS (
		SELECT days_used, COUNT(*) AS frequency
		FROM t2
		GROUP BY 1
	),
	t4 AS (
		SELECT SUM(frequency) AS total
		FROM t3
	)
	SELECT days_used, frequency, frequency / t4.total AS percentage
	FROM t3, t4
	ORDER BY days_used ASC;
	`
)

func (s *Users) Frequency() (*UsersFrequencyFetcher, error) {
	_, dateRangeCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	query := sqlf.Sprintf(frequencyQuery, dateRangeCond)

	return &UsersFrequencyFetcher{
		db:        s.DB,
		dateRange: s.DateRange,
		query:     query,
		group:     "Users:Frequency",
	}, nil
}

type UsersFrequencyFetcher struct {
	db        database.DB
	group     string
	dateRange string
	query     *sqlf.Query
}

type UsersFrequencyNodeData struct {
	DaysUsed   int32
	Frequency  int32
	Percentage float64
}

type UsersFrequencyNode struct {
	Data UsersFrequencyNodeData
}

func (n *UsersFrequencyNode) DaysUsed() int32 { return n.Data.DaysUsed }

func (n *UsersFrequencyNode) Frequency() int32 { return n.Data.Frequency }

func (n *UsersFrequencyNode) Percentage() float64 { return n.Data.Percentage }

func (f *UsersFrequencyFetcher) GetFrequencies(ctx context.Context, cache bool) ([]*UsersFrequencyNode, error) {
	cacheKey := "Users:Frequencies"
	if cache == true {
		if nodes, err := getArrayFromCache[UsersFrequencyNode](cacheKey); err == nil {
			return nodes, nil
		}
	}

	rows, err := f.db.QueryContext(ctx, f.query.Query(sqlf.PostgresBindVar), f.query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := make([]*UsersFrequencyNode, 0)
	for rows.Next() {
		var data UsersFrequencyNodeData

		if err := rows.Scan(&data.DaysUsed, &data.Frequency, &data.Percentage); err != nil {
			return nil, err
		}

		nodes = append(nodes, &UsersFrequencyNode{data})
	}

	if _, err := setArrayToCache(cacheKey, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (u *Users) CacheAll(ctx context.Context) error {
	activityFetcher, err := u.Activity()
	if err != nil {
		return err
	}

	if _, err := activityFetcher.GetNodes(ctx, false); err != nil {
		return err
	}

	if _, err := activityFetcher.GetSummary(ctx, false); err != nil {
		return err
	}

	frequenciesFetcher, err := u.Frequency()
	if err != nil {
		return err
	}

	if _, err := frequenciesFetcher.GetFrequencies(ctx, false); err != nil {
		return err
	}
	return nil
}

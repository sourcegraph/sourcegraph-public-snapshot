package adminanalytics

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Users struct {
	Ctx       context.Context
	DateRange string
	Grouping  string
	DB        database.DB
	Cache     bool
}

func (s *Users) Activity() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, s.Grouping, []string{})
	if err != nil {
		return nil, err
	}

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Users:Activity",
		cache:        s.Cache,
	}, nil
}

const (
	frequencyQuery = `
	WITH user_days_used AS (
        SELECT
            CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END AS user_id,
            COUNT(DISTINCT DATE(TIMEZONE('UTC', timestamp))) AS days_used
        FROM event_logs
		LEFT OUTER JOIN users ON users.id = event_logs.user_id
        WHERE
            DATE(timestamp) %s
            AND (%s)
        GROUP BY 1
    ),
    days_used_frequency AS (
        SELECT days_used, COUNT(*) AS frequency
        FROM user_days_used
        GROUP BY 1
    ),
    days_used_total_frequency AS (
        SELECT
            days_used_frequency.days_used,
            SUM(more_days_used_frequency.frequency) AS frequency
        FROM days_used_frequency
            LEFT JOIN days_used_frequency AS more_days_used_frequency
            ON more_days_used_frequency.days_used >= days_used_frequency.days_used
        GROUP BY 1
    ),
    max_days_used_total_frequency AS (
        SELECT MAX(frequency) AS max_frequency
        FROM days_used_total_frequency
    )
    SELECT
        days_used,
        frequency,
        frequency * 100.00 / COALESCE(max_frequency, 1) AS percentage
    FROM days_used_total_frequency, max_days_used_total_frequency
    ORDER BY 1 ASC;
	`
)

func (f *Users) Frequencies(ctx context.Context) ([]*UsersFrequencyNode, error) {
	cacheKey := fmt.Sprintf("Users:%s:%s", "Frequencies", f.DateRange)
	if f.Cache {
		if nodes, err := getArrayFromCache[UsersFrequencyNode](cacheKey); err == nil {
			return nodes, nil
		}
	}

	_, dateRangeCond, err := makeDateParameters(f.DateRange, f.Grouping, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(frequencyQuery, dateRangeCond, sqlf.Join(getDefaultConds(), ") AND ("))

	rows, err := f.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)

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

	if err := setArrayToCache(cacheKey, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

type UsersFrequencyNodeData struct {
	DaysUsed   float64
	Frequency  float64
	Percentage float64
}

type UsersFrequencyNode struct {
	Data UsersFrequencyNodeData
}

func (n *UsersFrequencyNode) DaysUsed() float64 { return n.Data.DaysUsed }

func (n *UsersFrequencyNode) Frequency() float64 { return n.Data.Frequency }

func (n *UsersFrequencyNode) Percentage() float64 { return n.Data.Percentage }

const (
	mauQuery = `
	SELECT
		TO_CHAR(TIMEZONE('UTC', timestamp), 'YYYY-MM') AS date,
		COUNT(DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) AS count
	FROM event_logs
	LEFT OUTER JOIN users ON users.id = event_logs.user_id
	WHERE
		timestamp BETWEEN %s AND %s
    AND (%s)
	GROUP BY 1
	ORDER BY 1 ASC
	`
)

func (f *Users) MonthlyActiveUsers(ctx context.Context) ([]*MonthlyActiveUsersRow, error) {
	cacheKey := fmt.Sprintf("Users:%s", "MAU")
	if f.Cache {
		if nodes, err := getArrayFromCache[MonthlyActiveUsersRow](cacheKey); err == nil {
			return nodes, nil
		}
	}

	from, to := getTimestamps(2) // go back 2 months

	query := sqlf.Sprintf(mauQuery, from, to, sqlf.Join(getDefaultConds(), ") AND ("))

	rows, err := f.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]*MonthlyActiveUsersRow, 0, 3)
	for rows.Next() {
		var data MonthlyActiveUsersRowData

		if err := rows.Scan(&data.Date, &data.Count); err != nil {
			return nil, err
		}

		nodes = append(nodes, &MonthlyActiveUsersRow{data})
	}

	if err := setArrayToCache(cacheKey, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

type MonthlyActiveUsersRowData struct {
	Date  string
	Count float64
}

type MonthlyActiveUsersRow struct {
	Data MonthlyActiveUsersRowData
}

func (n *MonthlyActiveUsersRow) Date() string { return n.Data.Date }

func (n *MonthlyActiveUsersRow) Count() float64 { return n.Data.Count }

func (u *Users) CacheAll(ctx context.Context) error {
	activityFetcher, err := u.Activity()
	if err != nil {
		return err
	}

	if _, err := activityFetcher.Nodes(ctx); err != nil {
		return err
	}

	if _, err := activityFetcher.Summary(ctx); err != nil {
		return err
	}

	if _, err := u.Frequencies(ctx); err != nil {
		return err
	}

	if _, err := u.MonthlyActiveUsers(ctx); err != nil {
		return err
	}
	return nil
}

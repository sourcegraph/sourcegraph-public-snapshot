package adminanalytics

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Users struct {
	DateRange string
	DB        database.DB
	Cache     bool
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
		cache:        s.Cache,
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

func (f *Users) Frequencies(ctx context.Context) ([]*UsersFrequencyNode, error) {
	_, dateRangeCond, err := makeDateParameters(f.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	query := sqlf.Sprintf(frequencyQuery, dateRangeCond)
	cacheKey := fmt.Sprintf("Users:%s:%s", f.DateRange, "Frequencies")
	if f.Cache == true {
		if nodes, err := getArrayFromCache[UsersFrequencyNode](cacheKey); err == nil {
			return nodes, nil
		}
	}

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

	if _, err := setArrayToCache(cacheKey, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
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

var (
	avgUsersByPeriodQuery = `
	WITH daus AS (
		SELECT
			DATE_TRUNC('day', event_logs.timestamp) AS day,
			COUNT(DISTINCT event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM
			users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
		WHERE
			event_logs.timestamp >= DATE_TRUNC('day', NOW() - CAST(%[1]v || ' days' AS INTERVAL))
		GROUP BY
			day
	),
	waus AS (
		SELECT
			DATE_TRUNC('week', event_logs.timestamp) AS week,
			COUNT(DISTINCT event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM
			users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
		WHERE
			event_logs.timestamp >= DATE_TRUNC('week', NOW() - CAST(%[1]v || ' days' AS INTERVAL))
		GROUP BY
			week
	),
	maus AS (
		SELECT
			DATE_TRUNC('month', event_logs.timestamp) AS month,
			COUNT(DISTINCT event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM
			users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
		WHERE
			event_logs.timestamp >= DATE_TRUNC('month', NOW() - CAST(%[1]v || ' days' AS INTERVAL))
		GROUP BY
			month
	)
	SELECT
		ROUND(SUM(daus.total_count) / COUNT(day)) AS avg_total_count,
		ROUND(SUM(daus.unique_users) / COUNT(day)) AS avg_unique_users,
		ROUND(SUM(daus.registered_users) / COUNT(day)) AS avg_registered_users
	FROM
		daus
	UNION ALL
	SELECT
		ROUND(SUM(waus.total_count) / COUNT(week)) AS avg_total_count,
		ROUND(SUM(waus.unique_users) / COUNT(week)) AS av_unique_users,
		ROUND(SUM(waus.registered_users) / COUNT(week)) AS av_registered_users
	FROM
		waus
	UNION ALL
	SELECT
		ROUND(SUM(maus.total_count) / COUNT(month)) AS avg_total_count,
		ROUND(SUM(maus.unique_users) / COUNT(month)) AS av_unique_users,
		ROUND(SUM(maus.registered_users) / COUNT(month)) AS av_registered_users
	FROM
		maus
	`
)

func (s *Users) Summary(ctx context.Context) (*UsersSummary, error) {
	now := time.Now()
	from, err := getFromDate(s.DateRange, now)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("Users:%s:%s", s.DateRange, "Summary")

	if s.Cache == true {
		if summary, err := getItemFromCache[UsersSummary](cacheKey); err == nil {
			return summary, nil
		}
	}
	days := int(now.Sub(from).Hours()/24) + 1
	query := sqlf.Sprintf(fmt.Sprintf(avgUsersByPeriodQuery, days))
	rows, err := s.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var summaryData UsersSummaryData
	for i := 0; i < 3; i++ {
		rows.Next()
		var data AnalyticsSummaryData
		if err := rows.Scan(&data.TotalCount, &data.TotalUniqueUsers, &data.TotalRegisteredUsers); err != nil {
			return nil, err
		}
		if i == 0 {
			summaryData.AvgDAU = AnalyticsSummary{data}
		} else if i == 1 {
			summaryData.AvgWAU = AnalyticsSummary{data}
		} else {
			summaryData.AvgMAU = AnalyticsSummary{data}
		}
	}

	summary := UsersSummary{summaryData}

	if _, err := setItemToCache(cacheKey, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

type UsersSummaryData struct {
	AvgDAU AnalyticsSummary
	AvgWAU AnalyticsSummary
	AvgMAU AnalyticsSummary
}

type UsersSummary struct {
	Data UsersSummaryData
}

func (s *UsersSummary) AvgDAU() *AnalyticsSummary { return &s.Data.AvgDAU }
func (s *UsersSummary) AvgWAU() *AnalyticsSummary { return &s.Data.AvgWAU }
func (s *UsersSummary) AvgMAU() *AnalyticsSummary { return &s.Data.AvgMAU }

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

	if _, err := u.Summary(ctx); err != nil {
		return err
	}
	return nil
}

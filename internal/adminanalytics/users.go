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

func (s *Users) Activity() (*AnalyticsFetcher, error) {
	nodesQuery, summaryQuery, err := makeEventLogsQueries(s.DateRange, []string{})
	if err != nil {
		return nil, err
	}

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
	cacheKey := fmt.Sprintf("Users:%s:%s", "Frequencies", f.DateRange)
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

var (
	avgUsersByPeriodQuery = `
	WITH daus AS (
		SELECT
			DATE_TRUNC('day', timestamp) AS day,
			0 AS total_count,
			COUNT(DISTINCT anonymous_user_id) AS unique_users,
			COUNT(DISTINCT user_id) FILTER (
				WHERE
					user_id != 0
			) AS registered_users
		FROM
			event_logs
			WHERE timestamp BETWEEN '%[1]v' AND '%[2]v'
		GROUP BY
			day
	),
	waus AS (
		SELECT
			DATE_TRUNC('week', timestamp) AS week,
			0 AS total_count,
			COUNT(DISTINCT anonymous_user_id) AS unique_users,
			COUNT(DISTINCT user_id) FILTER (
				WHERE
					user_id != 0
			) AS registered_users
		FROM
			event_logs
			WHERE timestamp BETWEEN '%[1]v' AND '%[2]v'
		GROUP BY
			week
	),
	maus AS (
		SELECT
			DATE_TRUNC('month', timestamp) AS month,
			0 AS total_count,
			COUNT(DISTINCT anonymous_user_id) AS unique_users,
			COUNT(DISTINCT user_id) FILTER (
				WHERE
					user_id != 0
			) AS registered_users
		FROM
			event_logs
			WHERE timestamp BETWEEN '%[1]v' AND '%[2]v'
		GROUP BY
			month
	)
	SELECT
		ROUND(sum_total_count / total_days)::int AS avg_total_count,
		ROUND(sum_unique_users / total_days)::int AS avg_unique_users,
		ROUND(sum_registered_users / total_days)::int AS avg_registered_users
	FROM
		(
			SELECT
				EXTRACT(
					EPOCH
					FROM(
						   '%[2]v' :: timestamp - '%[1]v' :: timestamp
						)
				) * 1.0 / 60 / 60 / 24 as total_days,
				SUM(daus.total_count) AS sum_total_count,
				SUM(daus.unique_users) AS sum_unique_users,
				SUM(daus.registered_users) AS sum_registered_users
			FROM
				daus
		) AS f
	UNION ALL
	SELECT
		ROUND(sum_total_count / total_weeks)::int AS avg_total_count,
		ROUND(sum_unique_users / total_weeks)::int AS avg_unique_users,
		ROUND(sum_registered_users / total_weeks)::int AS avg_registered_users
	FROM
		(
			SELECT
				EXTRACT(
					EPOCH
					FROM(
							'%[2]v' :: timestamp - '%[1]v' :: timestamp
						)
				) * 1.0 / 60 / 60 / 24 / 7 as total_weeks,
				SUM(waus.total_count) AS sum_total_count,
				SUM(waus.unique_users) AS sum_unique_users,
				SUM(waus.registered_users) AS sum_registered_users
			FROM
				waus
		) AS f
	UNION ALL
	SELECT
		ROUND(sum_total_count / total_months)::int AS avg_total_count,
		ROUND(sum_unique_users / total_months)::int AS avg_unique_users,
		ROUND(sum_registered_users / total_months)::int AS avg_registered_users
	FROM
		(
			SELECT
				EXTRACT(
					EPOCH
					FROM(
							'%[2]v' :: timestamp - '%[1]v' :: timestamp
						)
				) * 1.0 / 60 / 60 / 24 / 30 as total_months,
				SUM(maus.total_count) AS sum_total_count,
				SUM(maus.unique_users) AS sum_unique_users,
				SUM(maus.registered_users) AS sum_registered_users
			FROM
				maus
		) AS f
	`
)

func (s *Users) Summary(ctx context.Context) (*UsersSummary, error) {
	now := time.Now()
	from, err := getFromDate(s.DateRange, now)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("Users:%s:%s", "Summary", s.DateRange)

	if s.Cache == true {
		if summary, err := getItemFromCache[UsersSummary](cacheKey); err == nil {
			return summary, nil
		}
	}
	query := sqlf.Sprintf(fmt.Sprintf(avgUsersByPeriodQuery, from.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05")))
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

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
		SELECT user_id AS user_id, COUNT(date) AS days_used
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
		FLOOR(EXTRACT(EPOCH FROM('%[2]v'::timestamp - timestamp)) * 1.0 / 60 / 60 / 24) as day,
		COUNT(DISTINCT anonymous_user_id) AS total_count
		FROM
			event_logs
			WHERE timestamp BETWEEN '%[1]v' AND '%[2]v'
		GROUP BY
			day
	),
	waus AS (
		SELECT
			FLOOR(EXTRACT(EPOCH FROM('%[2]v'::timestamp - timestamp)) * 1.0 / 60 / 60 / 24 / 7) as week,
			COUNT(DISTINCT anonymous_user_id) AS total_count
		FROM
			event_logs
			WHERE timestamp BETWEEN '%[1]v' AND '%[2]v'
		GROUP BY
			week
	),
	maus AS (
		SELECT
			FLOOR(EXTRACT(EPOCH FROM('%[2]v'::timestamp - timestamp)) * 1.0 / 60 / 60 / 24 / 30) as month,
			COUNT(DISTINCT anonymous_user_id) AS total_count
		FROM
			event_logs
			WHERE timestamp BETWEEN '%[1]v' AND '%[2]v'
		GROUP BY
			month
	)

	SELECT
		'DAU' AS metric,
		ROUND(sum_total_count / total_days)::int AS avg_total_count
	FROM
		(
			SELECT
				EXTRACT(EPOCH FROM('%[2]v'::timestamp - '%[1]v' :: timestamp)) * 1.0 / 60 / 60 / 24 as total_days,
				SUM(daus.total_count) AS sum_total_count
			FROM
				daus
		) AS f

	UNION ALL

	SELECT
		'WAU' AS metric,
		CASE
			WHEN total_weeks > 1 THEN ROUND(sum_total_count / total_weeks)::int
			ELSE sum_total_count::int
		END AS avg_total_count
	FROM
		(
			SELECT
				EXTRACT(EPOCH FROM('%[2]v'::timestamp - '%[1]v' :: timestamp)) * 1.0 / 60 / 60 / 24 / 7 as total_weeks,
				SUM(waus.total_count) AS sum_total_count
			FROM
				waus
		) AS f

	UNION ALL

	SELECT
		'MAU' AS metric,
		CASE
			WHEN total_months > 1 THEN ROUND(sum_total_count / total_months)::int
			ELSE sum_total_count::int
		END AS avg_total_count
	FROM
		(
			SELECT
				EXTRACT(EPOCH FROM('%[2]v'::timestamp - '%[1]v' :: timestamp)) * 1.0 / 60 / 60 / 24 / 30 as total_months,
				SUM(maus.total_count) AS sum_total_count
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

		var totalCount float64
		var metric string
		if err := rows.Scan(&metric, &totalCount); err != nil {
			return nil, err
		}

		if metric == "DAU" {
			summaryData.AvgDAU = totalCount
		} else if metric == "WAU" {
			summaryData.AvgWAU = totalCount
		} else if metric == "MAU" {
			summaryData.AvgMAU = totalCount
		}
	}

	summary := UsersSummary{summaryData}

	if _, err := setItemToCache(cacheKey, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

type UsersSummaryData struct {
	AvgDAU float64
	AvgWAU float64
	AvgMAU float64
}

type UsersSummary struct {
	Data UsersSummaryData
}

func (s *UsersSummary) AvgDAU() float64 { return s.Data.AvgDAU }
func (s *UsersSummary) AvgWAU() float64 { return s.Data.AvgWAU }
func (s *UsersSummary) AvgMAU() float64 { return s.Data.AvgMAU }

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

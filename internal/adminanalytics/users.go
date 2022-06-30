package adminanalytics

import (
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Users struct {
	DateRange string
	DB        database.DB
}

func (s *Users) Activity() (*AnalyticsFetcher, error) {
	dateSelectParam, dateRangeCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	nodesQuery := sqlf.Sprintf(`
		SELECT %s AS date,
			COUNT(DISTINCT event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
		WHERE event_logs.timestamp %s
		GROUP BY date
	`, dateSelectParam, dateRangeCond)

	summaryQuery := sqlf.Sprintf(`
		SELECT
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			RIGHT JOIN event_logs ON users.id = event_logs.user_id
		WHERE event_logs.timestamp %s
	`, dateRangeCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Users:Activity",
	}, nil
}

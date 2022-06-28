package adminanalytics

import (
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Search struct {
	DateRange string
	DB        database.DB
}

func (s *Search) Searches() (*AnalyticsFetcher, error) {
	dateSelectParam, dateRangeCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	nodesQuery := sqlf.Sprintf(`
		SELECT %s AS date,
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('SearchResultsQueried')
		WHERE event_logs.timestamp %s
		GROUP BY date
	`, dateSelectParam, dateRangeCond)

	summaryQuery := sqlf.Sprintf(`
		SELECT
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('SearchResultsQueried')
		WHERE event_logs.timestamp %s
	`, dateRangeCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:Searches",
	}, nil
}

func (s *Search) FileViews() (*AnalyticsFetcher, error) {
	dateSelectParam, dateRangeCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	nodesQuery := sqlf.Sprintf(`
		SELECT %s AS date,
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('ViewBlob')
		WHERE event_logs.timestamp %s
		GROUP BY date
	`, dateSelectParam, dateRangeCond)

	summaryQuery := sqlf.Sprintf(`
		SELECT
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('ViewBlob')
		WHERE event_logs.timestamp %s
	`, dateRangeCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:FileViews",
	}, nil
}

func (s *Search) FileOpens() (*AnalyticsFetcher, error) {
	dateSelectParam, dateRangeCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	// TODO: add other open-in-ide events for other IDE plugins
	nodesQuery := sqlf.Sprintf(`
		SELECT %s AS date,
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('GoToCodeHostClicked', 'vscode.open.file')
		WHERE event_logs.timestamp %s
		GROUP BY date
	`, dateSelectParam, dateRangeCond)

	summaryQuery := sqlf.Sprintf(`
		SELECT
			COUNT(event_logs.*) AS total_count,
			COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
			COUNT(DISTINCT users.id) AS registered_users
		FROM users
			JOIN event_logs ON users.id = event_logs.user_id
			AND event_logs.name IN ('GoToCodeHostClicked', 'vscode.open.file')
		WHERE event_logs.timestamp %s
	`, dateRangeCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:FileOpens",
	}, nil
}

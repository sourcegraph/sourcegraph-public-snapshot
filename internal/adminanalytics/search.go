package adminanalytics

import (
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Search struct {
	DateRange string
	DB        database.DB
}

var eventLogsNodesQuery = `
SELECT
	%s AS date,
	COUNT(event_logs.*) AS total_count,
	COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
	COUNT(DISTINCT users.id) AS registered_users
FROM
	users
	RIGHT JOIN event_logs ON users.id = event_logs.user_id
WHERE event_logs.anonymous_user_id <> 'backend'
	AND event_logs.timestamp %s
	AND event_logs.name IN (%s)
GROUP BY date
`

var eventLogsSummaryQuery = `
SELECT
	COUNT(event_logs.*) AS total_count,
	COUNT(DISTINCT event_logs.anonymous_user_id) AS unique_users,
	COUNT(DISTINCT users.id) AS registered_users
FROM
	users
	RIGHT JOIN event_logs ON users.id = event_logs.user_id
WHERE
	event_logs.anonymous_user_id <> 'backend'
	AND event_logs.timestamp %s
	AND event_logs.name IN (%s)
`

func (s *Search) Searches() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	eventsCond := makeStringsInExpression([]string {"SearchResultsQueried"})
	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, dateBetweenCond, eventsCond)
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, dateBetweenCond, eventsCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:Searches",
	}, nil
}

func (s *Search) FileViews() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	eventsCond := makeStringsInExpression([]string {"ViewBlob"})
	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, dateBetweenCond, eventsCond)
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, dateBetweenCond, eventsCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:FileViews",
	}, nil
}

func (s *Search) FileOpens() (*AnalyticsFetcher, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(s.DateRange, "event_logs.timestamp")
	if err != nil {
		return nil, err
	}
	eventsCond := makeStringsInExpression([]string {
		"GoToCodeHostClicked",
		"vscode.open.file",
		"openInAtom.open.file",
		"openineditor.open.file",
		"openInWebstorm.open.file",
	})
	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, dateBetweenCond, eventsCond)
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, dateBetweenCond, eventsCond)

	return &AnalyticsFetcher{
		db:           s.DB,
		dateRange:    s.DateRange,
		nodesQuery:   nodesQuery,
		summaryQuery: summaryQuery,
		group:        "Search:FileOpens",
	}, nil
}

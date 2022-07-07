package adminanalytics

import (
	"errors"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
)

var (
	LastThreeMonths = "LAST_THREE_MONTHS"
	LastMonth       = "LAST_MONTH"
	LastWeek        = "LAST_WEEK"
)

func makeStringsInExpression(values []string) *sqlf.Query {
	var conds []*sqlf.Query
	for _, value := range values {
		conds = append(conds, sqlf.Sprintf("%s", value))
	}
	return sqlf.Join(conds, ",")
}

func makeDateParameters(dateRange string, dateColumnName string) (*sqlf.Query, *sqlf.Query, error) {
	now := time.Now()
	from, err := getFromDate(dateRange, now)
	if err != nil {
		return nil, nil, err
	}
	var groupBy string

	if dateRange == LastThreeMonths {
		groupBy = "week"
	} else if dateRange == LastMonth {
		groupBy = "day"
	} else if dateRange == LastWeek {
		groupBy = "day"
	} else {
		return nil, nil, errors.New("Invalid date range")
	}

	return sqlf.Sprintf(fmt.Sprintf(`DATE_TRUNC('%s', %s::date)`, groupBy, dateColumnName)), sqlf.Sprintf(`BETWEEN %s AND %s`, from.Format(time.RFC3339), now.Format(time.RFC3339)), nil
}

func getFromDate(dateRange string, now time.Time) (time.Time, error) {
	if dateRange == LastThreeMonths {
		return now.AddDate(0, -3, 0), nil
	} else if dateRange == LastMonth {
		return now.AddDate(0, -1, 0), nil
	} else if dateRange == LastWeek {
		return now.AddDate(0, 0, -7), nil
	}

	return now, errors.New("Invalid date range")
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

func makeEventLogsQueries(dateRange string, events []string) (*sqlf.Query, *sqlf.Query, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(dateRange, "event_logs.timestamp")
	if err != nil {
		return nil, nil, err
	}

	eventsCond := makeStringsInExpression(events)
	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, dateBetweenCond, eventsCond)
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, dateBetweenCond, eventsCond)

	return nodesQuery, summaryQuery, nil
}

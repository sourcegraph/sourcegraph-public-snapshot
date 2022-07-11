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
	COUNT(*) AS total_count,
	COUNT(DISTINCT anonymous_user_id) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
%s
GROUP BY date
`

var eventLogsSummaryQuery = `
SELECT
	COUNT(*) AS total_count,
	COUNT(DISTINCT anonymous_user_id) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
%s
`

func makeEventLogsQueries(dateRange string, events []string) (*sqlf.Query, *sqlf.Query, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(dateRange, "timestamp")
	if err != nil {
		return nil, nil, err
	}

	conds := []*sqlf.Query{
		sqlf.Sprintf("anonymous_user_id <> 'backend'"),
		sqlf.Sprintf("timestamp %s", dateBetweenCond),
	}

	if len(events) > 0 {
		var eventNames []*sqlf.Query
		for _, name := range events {
			eventNames = append(eventNames, sqlf.Sprintf("%s", name))
		}
		conds = append(conds, sqlf.Sprintf("name IN (%s)", sqlf.Join(eventNames, ",")))
	}

	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, sqlf.Sprintf("WHERE %s", sqlf.Join(conds, " AND ")))
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, sqlf.Sprintf("WHERE %s", sqlf.Join(conds, " AND ")))

	return nodesQuery, summaryQuery, nil
}

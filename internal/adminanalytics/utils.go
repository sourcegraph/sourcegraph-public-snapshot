package adminanalytics

import (
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	LastThreeMonths = "LAST_THREE_MONTHS"
	LastMonth       = "LAST_MONTH"
	LastWeek        = "LAST_WEEK"
	Daily           = "DAILY"
	Weekly          = "WEEKLY"
	timeNow         = time.Now
)

func makeDateParameters(dateRange string, grouping string, dateColumnName string) (*sqlf.Query, *sqlf.Query, error) {
	now := time.Now()
	from, err := getFromDate(dateRange, now)
	if err != nil {
		return nil, nil, err
	}
	var groupBy string

	if grouping == Weekly {
		groupBy = "week"
	} else if grouping == Daily {
		groupBy = "day"
	} else {
		return nil, nil, errors.New("Invalid groupBy")
	}

	return sqlf.Sprintf(fmt.Sprintf(`DATE_TRUNC('%s', TIMEZONE('UTC', %s::date))`, groupBy, dateColumnName)), sqlf.Sprintf(`BETWEEN %s AND %s`, from.Format(time.RFC3339), now.Format(time.RFC3339)), nil
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

const eventLogsNodesQuery = `
SELECT
	%s AS date,
	COUNT(*) AS total_count,
	COUNT(DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
LEFT OUTER JOIN users ON users.id = event_logs.user_id
%s
GROUP BY date
`

const eventLogsSummaryQuery = `
SELECT
	COUNT(*) AS total_count,
	COUNT(DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
LEFT OUTER JOIN users ON users.id = event_logs.user_id
%s
`

func getDefaultConds() []*sqlf.Query {
	commonConds := database.BuildCommonUsageConds(&database.CommonUsageOptions{
		ExcludeSystemUsers:          true,
		ExcludeNonActiveUsers:       true,
		ExcludeSourcegraphAdmins:    true,
		ExcludeSourcegraphOperators: true,
	}, []*sqlf.Query{})

	return append(commonConds, sqlf.Sprintf("anonymous_user_id != 'backend'"))
}

func makeEventLogsQueries(dateRange string, grouping string, events []string, conditions ...*sqlf.Query) (*sqlf.Query, *sqlf.Query, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(dateRange, grouping, "timestamp")
	if err != nil {
		return nil, nil, err
	}

	conds := append(getDefaultConds(), sqlf.Sprintf("timestamp %s", dateBetweenCond))

	if len(conditions) > 0 {
		conds = append(conds, conditions...)
	}

	if len(events) > 0 {
		var eventNames []*sqlf.Query
		for _, name := range events {
			eventNames = append(eventNames, sqlf.Sprintf("%s", name))
		}
		conds = append(conds, sqlf.Sprintf("name IN (%s)", sqlf.Join(eventNames, ",")))
	}

	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, sqlf.Sprintf("WHERE (%s)", sqlf.Join(conds, ") AND (")))
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, sqlf.Sprintf("WHERE (%s)", sqlf.Join(conds, ") AND (")))

	return nodesQuery, summaryQuery, nil
}

// getTimestamps returns the start and end timestamps for the given number of months.
func getTimestamps(months int) (string, string) {
	now := timeNow().UTC()
	to := now.Format(time.RFC3339)
	prevMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -months, 0)
	from := prevMonth.Format(time.RFC3339)

	return from, to
}

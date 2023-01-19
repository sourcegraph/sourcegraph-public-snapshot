package adminanalytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	LastThreeMonths = "LAST_THREE_MONTHS"
	LastMonth       = "LAST_MONTH"
	LastWeek        = "LAST_WEEK"
	Daily           = "DAILY"
	Weekly          = "WEEKLY"
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

// find sourcegraph employee user ids to exclude (usually CEs)
var sgEmpUserIdsQuery = `
SELECT
  DISTINCT users.id AS user_id
FROM
  users INNER JOIN user_emails ON user_emails.user_id = users.id
WHERE
  (
    users.username ILIKE 'managed-%%'
		OR users.username ILIKE 'sourcegraph-management-%%'
		OR users.username = 'sourcegraph-admin'
  )
	AND user_emails.email ILIKE '%%@sourcegraph.com'
`

const employeeUserIdsCacheExpirySeconds = 300
const employeeUserIdsCacheKey = "sourcegraph_employee_user_ids"

func getSgEmpUserIDs(ctx context.Context, db database.DB, cache bool) ([]*int32, error) {
	if cache {
		if ids, err := getArrayFromCache[int32](employeeUserIdsCacheKey); err == nil {
			return ids, nil
		}
	}

	query := sqlf.Sprintf(sgEmpUserIdsQuery)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return []*int32{}, err
	}
	defer rows.Close()

	ids := make([]*int32, 0)
	for rows.Next() {
		var id int32

		if err := rows.Scan(&id); err != nil {
			return ids, err
		}

		ids = append(ids, &id)
	}

	cacheData, err := json.Marshal(ids)
	if err != nil {
		return ids, err
	}

	if _, err := setDataToCache(employeeUserIdsCacheKey, string(cacheData), employeeUserIdsCacheExpirySeconds); err != nil {
		return ids, err
	}

	return ids, nil
}

const eventLogsNodesQuery = `
SELECT
	%s AS date,
	COUNT(*) AS total_count,
	COUNT(DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
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
%s
`

func getDefaultConds(ctx context.Context, db database.DB, cache bool) ([]*sqlf.Query, error) {
	nonActiveUserEvents := []*sqlf.Query{}
	for _, name := range eventlogger.NonActiveUserEvents {
		nonActiveUserEvents = append(nonActiveUserEvents, sqlf.Sprintf("%s", name))
	}

	conds := []*sqlf.Query{
		sqlf.Sprintf("anonymous_user_id <> 'backend'"),
		sqlf.Sprintf("name NOT IN (%s)", sqlf.Join(nonActiveUserEvents, ", ")),
		sqlf.Sprintf(fmt.Sprintf(`NOT public_argument @> '{"%s": true}'`, database.EventLogsSourcegraphOperatorKey)), // Exclude Sourcegraph Operator user accounts
	}

	sgEmpUserIds, err := getSgEmpUserIDs(ctx, db, cache)
	if err != nil {
		return []*sqlf.Query{}, err
	}

	if len(sgEmpUserIds) > 0 {
		excludeUserIDs := []*sqlf.Query{}
		for _, userId := range sgEmpUserIds {
			excludeUserIDs = append(excludeUserIDs, sqlf.Sprintf("%d", userId))
		}

		conds = append(conds, sqlf.Sprintf("user_id NOT IN (%s)", sqlf.Join(excludeUserIDs, ", ")))
	}

	return conds, nil
}

func makeEventLogsQueries(ctx context.Context, db database.DB, cache bool, dateRange string, grouping string, events []string, conditions ...*sqlf.Query) (*sqlf.Query, *sqlf.Query, error) {
	dateTruncExp, dateBetweenCond, err := makeDateParameters(dateRange, grouping, "timestamp")
	if err != nil {
		return nil, nil, err
	}

	defaultConds, err := getDefaultConds(ctx, db, cache)
	if err != nil {
		return nil, nil, err
	}

	conds := append(defaultConds, sqlf.Sprintf("timestamp %s", dateBetweenCond))

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

	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dateTruncExp, sqlf.Sprintf("WHERE %s", sqlf.Join(conds, " AND ")))
	summaryQuery := sqlf.Sprintf(eventLogsSummaryQuery, sqlf.Sprintf("WHERE %s", sqlf.Join(conds, " AND ")))

	return nodesQuery, summaryQuery, nil
}

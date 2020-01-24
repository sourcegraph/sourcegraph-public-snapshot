package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const (
	integrationSource = "CODEHOSTINTEGRATION"
)

type eventLogs struct{}

// Event contains information needed for logging an event.
type Event struct {
	Name            string
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        json.RawMessage
	Source          string
	Timestamp       time.Time
}

func (*eventLogs) Insert(ctx context.Context, e *Event) error {
	argument := e.Argument
	if argument == nil {
		argument = json.RawMessage([]byte(`{}`))
	}

	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id, source, argument, version, timestamp) VALUES($1, $2, $3, $4, $5, $6, $7, $8)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Source,
		argument,
		version.Version(),
		e.Timestamp.UTC(),
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}

func (*eventLogs) getBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*types.Event, error) {
	q := sqlf.Sprintf("SELECT id, name, url, user_id, anonymous_user_id, source, argument, version, timestamp FROM event_logs %s", querySuffix)
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []*types.Event{}
	for rows.Next() {
		r := types.Event{}
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.UserID, &r.AnonymousUserID, &r.Source, &r.Argument, &r.Version, &r.Timestamp)
		if err != nil {
			return nil, err
		}
		events = append(events, &r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// GetAll gets all event logs in descending order of timestamp.
func (l *eventLogs) GetAll(ctx context.Context) ([]*types.Event, error) {
	return l.getBySQL(ctx, sqlf.Sprintf("ORDER BY timestamp DESC"))
}

// GetByUserID gets all event logs by a given user in descending order of timestamp.
func (l *eventLogs) GetByUserID(ctx context.Context, userID int32) ([]*types.Event, error) {
	return l.getBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d ORDER BY timestamp DESC", userID))
}

// CountByUserIDAndEventName gets a count of events logged by a given user and with a given event name.
func (l *eventLogs) CountByUserIDAndEventName(ctx context.Context, userID int32, name string) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND name = %s", userID, name))
}

// CountByUserIDAndEventNamePrefix gets a count of events logged by a given user and with a given event name prefix.
func (l *eventLogs) CountByUserIDAndEventNamePrefix(ctx context.Context, userID int32, namePrefix string) (int, error) {
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND name LIKE %s", userID, namePrefix+"%"))
}

// CountByUserIDAndEventNames gets a count of events logged by a given user that match a list of given event names.
func (l *eventLogs) CountByUserIDAndEventNames(ctx context.Context, userID int32, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	return l.countBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND name IN (%s)", userID, sqlf.Join(items, ",")))
}

// countBySQL gets a count of event logs.
func (*eventLogs) countBySQL(ctx context.Context, querySuffix *sqlf.Query) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM event_logs %s", querySuffix)
	r := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	var count int
	err := r.Scan(&count)
	return count, err
}

// MaxTimestampByUserID gets the max timestamp among event logs for a given user.
func (l *eventLogs) MaxTimestampByUserID(ctx context.Context, userID int32) (*time.Time, error) {
	return l.maxTimestampBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d", userID))
}

// MaxTimestampByUserIDAndSource gets the max timestamp among event logs for a given user and event source.
func (l *eventLogs) MaxTimestampByUserIDAndSource(ctx context.Context, userID int32, source string) (*time.Time, error) {
	return l.maxTimestampBySQL(ctx, sqlf.Sprintf("WHERE user_id = %d AND source = %s", userID, source))
}

// maxTimestampBySQL gets the max timestamp among event logs.
func (*eventLogs) maxTimestampBySQL(ctx context.Context, querySuffix *sqlf.Query) (*time.Time, error) {
	q := sqlf.Sprintf("SELECT MAX(timestamp) FROM event_logs %s", querySuffix)
	r := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)

	var t time.Time
	err := r.Scan(&dbutil.NullTime{Time: &t})
	if t.IsZero() {
		return nil, err
	}
	return &t, err
}

// UsageValue is a single count of usage for a time period starting on a given date.
type UsageValue struct {
	Start time.Time
	Count int
}

// UniqueUserCountType is the type of period in which to count unique users.
type UniqueUserCountType string

const (
	// Daily is used to get a count of unique daily active users.
	Daily UniqueUserCountType = "daily"
	// Weekly is used to get a count of unique weekly active users.
	Weekly UniqueUserCountType = "weekly"
	// Monthly is used to get a count of unique monthly active users.
	Monthly UniqueUserCountType = "monthly"
)

// CountUniquesOptions provides options for counting unique users.
type CountUniquesOptions struct {
	// If true, only include registered users. Otherwise, include all users.
	RegisteredOnly bool
	// If true, only include code host integration users. Otherwise, include all users.
	IntegrationOnly bool
}

// CountUniquesPerPeriod provides a count of unique active users in a given time span, broken up into periods of a given type.
// Returns an array array of length `periods`, with one entry for each period in the time span.
func (l *eventLogs) CountUniquesPerPeriod(ctx context.Context, periodType UniqueUserCountType, startDate time.Time, periods int, opt *CountUniquesOptions) ([]UsageValue, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		if opt.RegisteredOnly {
			conds = append(conds, sqlf.Sprintf("user_id > 0"))
		}
		if opt.IntegrationOnly {
			conds = append(conds, sqlf.Sprintf("source = %s", integrationSource))
		}
	}
	switch periodType {
	case "daily":
		return l.countUniquesPerPeriodBySQL(ctx, sqlf.Sprintf("day"), sqlf.Sprintf("DATE_TRUNC('day', timestamp)"), startDate, startDate.AddDate(0, 0, periods), conds)
	case "weekly":
		return l.countUniquesPerPeriodBySQL(ctx, sqlf.Sprintf("week"), sqlf.Sprintf("DATE_TRUNC('week', timestamp + '1 day'::interval) - '1 day'::interval"), startDate, startDate.AddDate(0, 0, 7*periods), conds)
	case "monthly":
		return l.countUniquesPerPeriodBySQL(ctx, sqlf.Sprintf("month"), sqlf.Sprintf("DATE_TRUNC('month', timestamp)"), startDate, startDate.AddDate(0, periods, 0), conds)
	}
	return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
}

func (l *eventLogs) countUniquesPerPeriodBySQL(ctx context.Context, interval, period *sqlf.Query, startDate, endDate time.Time, conds []*sqlf.Query) ([]UsageValue, error) {
	allPeriods := sqlf.Sprintf("SELECT generate_series((%s)::timestamp, (%s)::timestamp,  (%s)::interval) AS period", startDate, endDate, sqlf.Sprintf("'1 %s'", interval))
	usersByPeriod := sqlf.Sprintf(`SELECT (%s) AS period, COUNT(DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) AS count
		FROM event_logs
		WHERE (%s)
		GROUP BY period`, period, sqlf.Join(conds, ") AND ("))
	q := sqlf.Sprintf(`WITH all_periods AS (%s), users_by_period AS (%s)
		SELECT all_periods.period, COALESCE(count, 0)
		FROM all_periods
		LEFT OUTER JOIN users_by_period ON all_periods.period = (users_by_period.period)::timestamp
		ORDER BY period DESC`, allPeriods, usersByPeriod)
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := []UsageValue{}
	for rows.Next() {
		var v UsageValue
		err := rows.Scan(&v.Start, &v.Count)
		if err != nil {
			return nil, err
		}
		v.Start = v.Start.UTC()
		counts = append(counts, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

// CountUniquesAll provides a count of unique active users in a given time span.
func (l *eventLogs) CountUniquesAll(ctx context.Context, startDate, endDate time.Time) (int, error) {
	return l.countUniquesBySQL(ctx, startDate, endDate, nil)
}

// CountUniquesByEventNamePrefix provides a count of unique active users in a given time span that logged an event with a given prefix.
func (l *eventLogs) CountUniquesByEventNamePrefix(ctx context.Context, startDate, endDate time.Time, namePrefix string) (int, error) {
	return l.countUniquesBySQL(ctx, startDate, endDate, sqlf.Sprintf("AND name LIKE %s ", namePrefix+"%"))
}

// CountUniquesByEventName provides a count of unique active users in a given time span that logged a given event.
func (l *eventLogs) CountUniquesByEventName(ctx context.Context, startDate, endDate time.Time, name string) (int, error) {
	return l.countUniquesBySQL(ctx, startDate, endDate, sqlf.Sprintf("AND name = %s", name))
}

// CountUniquesByEventNames provides a count of unique active users in a given time span that logged any event that matches a list of given event names
func (l *eventLogs) CountUniquesByEventNames(ctx context.Context, startDate, endDate time.Time, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	return l.countUniquesBySQL(ctx, startDate, endDate, sqlf.Sprintf("AND name IN (%s)", sqlf.Join(items, ",")))
}

func (*eventLogs) countUniquesBySQL(ctx context.Context, startDate, endDate time.Time, querySuffix *sqlf.Query) (int, error) {
	if querySuffix == nil {
		querySuffix = sqlf.Sprintf("")
	}
	q := sqlf.Sprintf(`SELECT COUNT(DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END)
		FROM event_logs
		WHERE (DATE(TIMEZONE('UTC'::text, timestamp)) >= %s) AND (DATE(TIMEZONE('UTC'::text, timestamp)) <= %s) %s`, startDate, endDate, querySuffix)
	r := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	var count int
	err := r.Scan(&count)
	return count, err
}

func (l *eventLogs) ListUniquesAll(ctx context.Context, startDate, endDate time.Time) ([]int32, error) {
	rows, err := dbconn.Global.QueryContext(ctx, `SELECT user_id
		FROM event_logs
		WHERE user_id > 0 AND DATE(TIMEZONE('UTC'::text, timestamp)) >= $1 AND DATE(TIMEZONE('UTC'::text, timestamp)) <= $2
		GROUP BY user_id`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	var users []int32
	defer rows.Close()
	for rows.Next() {
		var userID int32
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

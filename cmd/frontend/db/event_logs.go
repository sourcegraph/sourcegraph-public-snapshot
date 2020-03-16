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
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
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

// PeriodType is the type of period in which to count events and unique users.
type PeriodType string

const (
	// Daily is used to get a count of events or unique users within a day.
	Daily PeriodType = "daily"
	// Weekly is used to get a count of events or unique users within a week.
	Weekly PeriodType = "weekly"
	// Monthly is used to get a count of events or unique users within a month.
	Monthly PeriodType = "monthly"
)

// intervalByPeriodType is a map of generate_series step values by period type.
var intervalByPeriodType = map[PeriodType]*sqlf.Query{
	Daily:   sqlf.Sprintf("'1 day'"),
	Weekly:  sqlf.Sprintf("'1 week'"),
	Monthly: sqlf.Sprintf("'1 month'"),
}

// periodByPeriodType is a map of SQL fragments that produce a timestamp bucket by period
// type. This assumes the existence of a  field named `timestamp` in the enclosing query.
var periodByPeriodType = map[PeriodType]*sqlf.Query{
	Daily:   sqlf.Sprintf("DATE_TRUNC('day', timestamp)"),
	Weekly:  sqlf.Sprintf("DATE_TRUNC('week', timestamp + '1 day'::interval) - '1 day'::interval"),
	Monthly: sqlf.Sprintf("DATE_TRUNC('month', timestamp)"),
}

// calcStartDate calculates the the starting date of a number of periods given the period type.
// from the current time supplied as `now`. Returns a second false value if the period type is
// illegal.
func calcStartDate(now time.Time, periodType PeriodType, periods int) (time.Time, bool) {
	periodsAgo := periods - 1

	switch periodType {
	case Daily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -periodsAgo), true
	case Weekly:
		return timeutil.StartOfWeek(now, periodsAgo), true
	case Monthly:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -periodsAgo, 0), true
	}
	return time.Time{}, false
}

// calcEndDate calculates the the ending date of a number of periods given the period type.
// Returns a second false value if the period type is illegal.
func calcEndDate(startDate time.Time, periodType PeriodType, periods int) (time.Time, bool) {
	periodsAgo := periods - 1

	switch periodType {
	case Daily:
		return startDate.AddDate(0, 0, periodsAgo), true
	case Weekly:
		return startDate.AddDate(0, 0, 7*periodsAgo), true
	case Monthly:
		return startDate.AddDate(0, periodsAgo, 0), true
	}

	return time.Time{}, false
}

// CountUniqueUsersOptions provides options for counting unique users.
type CountUniqueUsersOptions struct {
	// If true, only include registered users. Otherwise, include all users.
	RegisteredOnly bool
	// If true, only include code host integration users. Otherwise, include all users.
	IntegrationOnly bool
	// If set, adds additional restrictions on the event types.
	EventFilters *EventFilterOptions
}

// EventFilterOptions provides options for filtering events.
type EventFilterOptions struct {
	// If set, only include events with a given prefix.
	ByEventNamePrefix string
	// If set, only include events with the given name.
	ByEventName string
	// If not empty, only include events that matche a list of given event names
	ByEventNames []string
	// Must be used with ByEventName
	//
	// If set, only include events that match a specified condition.
	ByEventNameWithCondition *sqlf.Query
}

// EventArgumentMatch provides the options for matching an event with
// a specific JSON value passed as an argument.
type EventArgumentMatch struct {
	// The name of the JSON key to match against.
	ArgumentName string
	// The actual value passed to the JSON key to match.
	ArgumentValue string
}

// CountUniqueUsersPerPeriod provides a count of unique active users in a given time span, broken up into periods of
// a given type. The value of `now` should be the current time in UTC. Returns an array of length `periods`, with one
// entry for each period in the time span.
func (l *eventLogs) CountUniqueUsersPerPeriod(ctx context.Context, periodType PeriodType, now time.Time, periods int, opt *CountUniqueUsersOptions) ([]UsageValue, error) {
	startDate, ok := calcStartDate(now, periodType, periods)
	if !ok {
		return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
	}

	endDate, ok := calcEndDate(startDate, periodType, periods)
	if !ok {
		return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
	}

	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		if opt.RegisteredOnly {
			conds = append(conds, sqlf.Sprintf("user_id > 0"))
		}
		if opt.IntegrationOnly {
			conds = append(conds, sqlf.Sprintf("source = %s", integrationSource))
		}
		if opt.EventFilters != nil {
			if opt.EventFilters.ByEventNamePrefix != "" {
				conds = append(conds, sqlf.Sprintf("name LIKE %s", opt.EventFilters.ByEventNamePrefix+"%"))
			}
			if opt.EventFilters.ByEventName != "" {
				conds = append(conds, sqlf.Sprintf("name = %s", opt.EventFilters.ByEventName))
			}
			if opt.EventFilters.ByEventNameWithCondition != nil {
				conds = append(conds, opt.EventFilters.ByEventNameWithCondition)
			}
			if len(opt.EventFilters.ByEventNames) > 0 {
				items := []*sqlf.Query{}
				for _, v := range opt.EventFilters.ByEventNames {
					items = append(items, sqlf.Sprintf("%s", v))
				}
				conds = append(conds, sqlf.Sprintf("name IN (%s)", sqlf.Join(items, ",")))
			}
		}
	}

	return l.countUniqueUsersPerPeriodBySQL(ctx, intervalByPeriodType[periodType], periodByPeriodType[periodType], startDate, endDate, conds)
}

// CountEventsPerPeriod provide a count of events in a given time span, broken up into periods of a given type.
// The value of `now` should be the current time in UTC.
func (l *eventLogs) CountEventsPerPeriod(ctx context.Context, periodType PeriodType, now time.Time, periods int, opt *EventFilterOptions) ([]UsageValue, error) {
	startDate, ok := calcStartDate(now, periodType, periods)
	if !ok {
		return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
	}

	endDate, ok := calcEndDate(startDate, periodType, periods)
	if !ok {
		return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
	}

	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		if opt.ByEventNamePrefix != "" {
			conds = append(conds, sqlf.Sprintf("name LIKE %s", opt.ByEventNamePrefix+"%"))
		}
		if opt.ByEventName != "" {
			conds = append(conds, sqlf.Sprintf("name = %s", opt.ByEventName))
		}
		if opt.ByEventNameWithCondition != nil {
			if opt.ByEventName == "" {
				return nil, fmt.Errorf("The ByEventNameWithCondition option must be used together with the ByEventName option.")
			}
			conds = append(conds, sqlf.Sprintf("name = %s", opt.ByEventName))
			conds = append(conds, opt.ByEventNameWithCondition)
		}
		if len(opt.ByEventNames) > 0 {
			items := []*sqlf.Query{}
			for _, v := range opt.ByEventNames {
				items = append(items, sqlf.Sprintf("%s", v))
			}
			conds = append(conds, sqlf.Sprintf("name IN (%s)", sqlf.Join(items, ",")))
		}
	}

	return l.countEventsPerPeriodBySQL(ctx, intervalByPeriodType[periodType], periodByPeriodType[periodType], startDate, endDate, conds)
}

// PercentileValue is a slice of Nth percentile values calculated from a field of events
// in a time period starting on a given date.
type PercentileValue struct {
	Start  time.Time
	Values []float64
}

// PercentilesPerPeriod calculates the given percentiles over a field of the event's arguments in a given time span,
// broken up into periods of a given type. The value of `now` should be the current time in UTC. Percentiles should
// be supplied as floats in the range `[0, 1)`, such that `[.5, .9, .99]` will return the 50th, 90th, and 99th
// percentile values. Returns an array of length `periods`, with one entry for each period in the time span. Each
// `PercentileValue` object in the result will contain exactly `len(percentiles)` values.
func (l *eventLogs) PercentilesPerPeriod(
	ctx context.Context,
	periodType PeriodType,
	now time.Time,
	periods int,
	field string,
	percentiles []float64,
	opt *EventFilterOptions,
) ([]PercentileValue, error) {
	if len(percentiles) == 0 {
		return nil, fmt.Errorf("expected at least one percentile value in query")
	}

	startDate, ok := calcStartDate(now, periodType, periods)
	if !ok {
		return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
	}

	endDate, ok := calcEndDate(startDate, periodType, periods)
	if !ok {
		return nil, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
	}

	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt != nil {
		if opt.ByEventNamePrefix != "" {
			conds = append(conds, sqlf.Sprintf("name LIKE %s", opt.ByEventNamePrefix+"%"))
		}
		if opt.ByEventName != "" {
			conds = append(conds, sqlf.Sprintf("name = %s", opt.ByEventName))
		}
		if len(opt.ByEventNames) > 0 {
			items := []*sqlf.Query{}
			for _, v := range opt.ByEventNames {
				items = append(items, sqlf.Sprintf("%s", v))
			}
			conds = append(conds, sqlf.Sprintf("name IN (%s)", sqlf.Join(items, ",")))
		}
	}

	return l.calculatePercentilesPerPeriodBySQL(ctx, intervalByPeriodType[periodType], periodByPeriodType[periodType], startDate, endDate, field, percentiles, conds)
}

func (l *eventLogs) countUniqueUsersPerPeriodBySQL(ctx context.Context, interval, period *sqlf.Query, startDate, endDate time.Time, conds []*sqlf.Query) ([]UsageValue, error) {
	return l.countPerPeriodBySQL(ctx, sqlf.Sprintf("DISTINCT CASE WHEN user_id = 0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END"), interval, period, startDate, endDate, conds)
}

func (l *eventLogs) countEventsPerPeriodBySQL(ctx context.Context, interval, period *sqlf.Query, startDate, endDate time.Time, conds []*sqlf.Query) ([]UsageValue, error) {
	return l.countPerPeriodBySQL(ctx, sqlf.Sprintf("*"), interval, period, startDate, endDate, conds)
}

func (l *eventLogs) countPerPeriodBySQL(ctx context.Context, countExpr, interval, period *sqlf.Query, startDate, endDate time.Time, conds []*sqlf.Query) ([]UsageValue, error) {
	allPeriods := sqlf.Sprintf("SELECT generate_series((%s)::timestamp, (%s)::timestamp, (%s)::interval) AS period", startDate, endDate, interval)
	countByPeriod := sqlf.Sprintf(`SELECT (%s) AS period, COUNT(%s) AS count
		FROM event_logs
		WHERE (%s)
		GROUP BY period`, period, countExpr, sqlf.Join(conds, ") AND ("))
	q := sqlf.Sprintf(`WITH all_periods AS (%s), count_by_period AS (%s)
		SELECT all_periods.period, COALESCE(count, 0)
		FROM all_periods
		LEFT OUTER JOIN count_by_period ON all_periods.period = (count_by_period.period)::timestamp
		ORDER BY period DESC`, allPeriods, countByPeriod)
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

func (l *eventLogs) calculatePercentilesPerPeriodBySQL(
	ctx context.Context,
	interval *sqlf.Query,
	period *sqlf.Query,
	startDate time.Time,
	endDate time.Time,
	field string,
	percentiles []float64,
	conds []*sqlf.Query,
) ([]PercentileValue, error) {
	countByPeriodExprs := []*sqlf.Query{}
	qExprs := []*sqlf.Query{}
	for i, p := range percentiles {
		name := fmt.Sprintf("p%d\n", i)
		// Note: we can't go directly from jsonb -> integer in postgres 9.6, so we
		// have to first cast it to a text type, and then to an integer to support
		// queries on older instances.
		countByPeriodExprs = append(countByPeriodExprs, sqlf.Sprintf(
			"percentile_cont(%d) WITHIN GROUP (ORDER BY (argument->%s)::text::integer) AS "+name,
			p,
			field,
		))

		qExprs = append(qExprs, sqlf.Sprintf("COALESCE("+name+", 0)"))
	}

	allPeriods := sqlf.Sprintf("SELECT generate_series((%s)::timestamp, (%s)::timestamp, (%s)::interval) AS period", startDate, endDate, interval)
	countByPeriod := sqlf.Sprintf(`SELECT (%s) AS period, %s
		FROM event_logs
		WHERE (%s)
		GROUP BY period`, period, sqlf.Join(countByPeriodExprs, ", "), sqlf.Join(conds, ") AND ("))
	q := sqlf.Sprintf(`WITH all_periods AS (%s), values_by_period AS (%s)
		SELECT all_periods.period, %s
		FROM all_periods
		LEFT OUTER JOIN values_by_period ON all_periods.period = (values_by_period.period)::timestamp
		ORDER BY period DESC`, allPeriods, countByPeriod, sqlf.Join(qExprs, ", "))
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []PercentileValue{}
	for rows.Next() {
		var v PercentileValue
		v.Values = make([]float64, len(percentiles))
		dest := []interface{}{&v.Start}
		for i := range v.Values {
			dest = append(dest, &v.Values[i])
		}
		err := rows.Scan(dest...)
		if err != nil {
			return nil, err
		}
		v.Start = v.Start.UTC()
		values = append(values, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// CountUniqueUsersAll provides a count of unique active users in a given time span.
func (l *eventLogs) CountUniqueUsersAll(ctx context.Context, startDate, endDate time.Time) (int, error) {
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, nil)
}

// CountUniqueUsersByEventNamePrefix provides a count of unique active users in a given time span that logged an event with a given prefix.
func (l *eventLogs) CountUniqueUsersByEventNamePrefix(ctx context.Context, startDate, endDate time.Time, namePrefix string) (int, error) {
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, sqlf.Sprintf("AND name LIKE %s ", namePrefix+"%"))
}

// CountUniqueUsersByEventName provides a count of unique active users in a given time span that logged a given event.
func (l *eventLogs) CountUniqueUsersByEventName(ctx context.Context, startDate, endDate time.Time, name string) (int, error) {
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, sqlf.Sprintf("AND name = %s", name))
}

// CountUniqueUsersByEventNames provides a count of unique active users in a given time span that logged any event that matches a list of given event names
func (l *eventLogs) CountUniqueUsersByEventNames(ctx context.Context, startDate, endDate time.Time, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	return l.countUniqueUsersBySQL(ctx, startDate, endDate, sqlf.Sprintf("AND name IN (%s)", sqlf.Join(items, ",")))
}

func (*eventLogs) countUniqueUsersBySQL(ctx context.Context, startDate, endDate time.Time, querySuffix *sqlf.Query) (int, error) {
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

func (l *eventLogs) ListUniqueUsersAll(ctx context.Context, startDate, endDate time.Time) ([]int32, error) {
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

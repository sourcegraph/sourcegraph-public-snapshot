package db

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const (
	webSource         = "WEB"
	integrationSource = "CODEHOSTINTEGRATION"
)

type eventLogs struct{}

// Event contains information needed for logging an event.
type Event struct {
	Name            string
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        string
	Source          string
	Timestamp       time.Time
}

func (*eventLogs) Insert(ctx context.Context, e *Event) error {
	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id, source, argument, version, timestamp) VALUES($1, $2, $3, $4, $5, $6, $7, $8)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Source,
		e.Argument,
		version.Version(),
		e.Timestamp,
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}

func (*eventLogs) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.Event, error) {
	rows, err := dbconn.Global.QueryContext(ctx, "SELECT id, name, url, user_id, anonymous_user_id, source, argument, version, timestamp FROM event_logs "+query, args...)
	if err != nil {
		return nil, err
	}
	responses := []*types.Event{}
	defer rows.Close()
	for rows.Next() {
		r := types.Event{}
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.UserID, &r.AnonymousUserID, &r.Source, &r.Argument, &r.Version, &r.Timestamp)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return responses, nil
}

// GetAll gets all event logs.
func (l *eventLogs) GetAll(ctx context.Context) ([]*types.Event, error) {
	return l.getBySQL(ctx, "ORDER BY timestamp DESC")
}

// GetByUserID gets all survey responses by a given user.
func (l *eventLogs) GetByUserID(ctx context.Context, userID int32) ([]*types.Event, error) {
	return l.getBySQL(ctx, "WHERE user_id=$1 ORDER BY created_at DESC", userID)
}

func (l *eventLogs) CountByUserIDAndEventName(ctx context.Context, userID int32, name string) (int, error) {
	return l.countBySQL(ctx, "WHERE user_id=$1 AND name = $2", userID, name)
}

func (l *eventLogs) CountByUserIDAndEventNamePrefix(ctx context.Context, userID int32, namePrefix string) (int, error) {
	return l.countBySQL(ctx, "WHERE user_id=$1 AND name LIKE $2 || '%'", userID, namePrefix)
}

func (l *eventLogs) CountByUserIDAndEventNames(ctx context.Context, userID int32, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	q := sqlf.Sprintf("WHERE user_id=%s AND name IN (%s)", userID, sqlf.Join(items, ","))
	return l.countBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

// Count returns the count of all survey responses.
func (*eventLogs) countBySQL(ctx context.Context, query string, args ...interface{}) (int, error) {
	r := dbconn.Global.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_logs "+query, args...)
	var count int
	err := r.Scan(&count)
	return count, err
}

func (l *eventLogs) MaxTimestampByUserID(ctx context.Context, userID int32) (*time.Time, error) {
	return l.getMaxTimestampBySQL(ctx, "WHERE user_id=$1", userID)
}

func (l *eventLogs) MaxTimestampByUserIDAndSource(ctx context.Context, userID int32, source string) (*time.Time, error) {
	return l.getMaxTimestampBySQL(ctx, "WHERE user_id=$1 AND source=$2", userID, source)
}

func (*eventLogs) getMaxTimestampBySQL(ctx context.Context, query string, args ...interface{}) (*time.Time, error) {
	r := dbconn.Global.QueryRowContext(ctx, "SELECT MAX(timestamp) FROM event_logs "+query, args...)

	var t *time.Time
	err := r.Scan(&t)
	return t, err
}

type UsageDatum struct {
	Start time.Time
	Count int
}

func (l *eventLogs) CountDAUs(ctx context.Context, startDate time.Time, dayPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "day", "DATE_TRUNC('day', timestamp)", "", startDate, startDate.AddDate(0, 0, dayPeriods))
}
func (l *eventLogs) CountRegisteredDAUs(ctx context.Context, startDate time.Time, dayPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "day", "DATE_TRUNC('day', timestamp)", "WHERE user_id > 0", startDate, startDate.AddDate(0, 0, dayPeriods))
}
func (l *eventLogs) CountIntegrationDAUs(ctx context.Context, startDate time.Time, dayPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "day", "DATE_TRUNC('day', timestamp)", "WHERE source = $3", startDate, startDate.AddDate(0, 0, dayPeriods), integrationSource)
}

func (l *eventLogs) CountWAUs(ctx context.Context, startDate time.Time, weekPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "week", "DATE_TRUNC('week', timestamp + '1 day'::interval) - '1 day'::interval", "", startDate, startDate.AddDate(0, 0, weekPeriods*7))
}
func (l *eventLogs) CountRegisteredWAUs(ctx context.Context, startDate time.Time, weekPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "week", "DATE_TRUNC('week', timestamp + '1 day'::interval) - '1 day'::interval", "WHERE user_id > 0", startDate, startDate.AddDate(0, 0, weekPeriods*7))
}
func (l *eventLogs) CountIntegrationWAUs(ctx context.Context, startDate time.Time, weekPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "week", "DATE_TRUNC('week', timestamp + '1 day'::interval) - '1 day'::interval", "WHERE source = $3", startDate, startDate.AddDate(0, 0, weekPeriods*7), integrationSource)
}

func (l *eventLogs) CountMAUs(ctx context.Context, startDate time.Time, monthPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "month", "DATE_TRUNC('month', timestamp)", "", startDate, startDate.AddDate(0, monthPeriods, 0))
}
func (l *eventLogs) CountRegisteredMAUs(ctx context.Context, startDate time.Time, monthPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "month", "DATE_TRUNC('month', timestamp)", "WHERE user_id > 0", startDate, startDate.AddDate(0, monthPeriods, 0))
}
func (l *eventLogs) CountIntegrationMAUs(ctx context.Context, startDate time.Time, monthPeriods int) ([]UsageDatum, error) {
	return l.countUniqueUsersBySQL(ctx, "month", "DATE_TRUNC('month', timestamp)", "WHERE source = $3", startDate, startDate.AddDate(0, monthPeriods, 0), integrationSource)
}

func (l *eventLogs) countUniqueUsersBySQL(ctx context.Context, interval string, period string, cond string, args ...interface{}) ([]UsageDatum, error) {
	rows, err := dbconn.Global.QueryContext(ctx, `
WITH
	all_periods AS (SELECT generate_series(($1)::timestamp, ($2)::timestamp, interval '1 `+interval+`') AS period),
	users_by_period AS (
		SELECT
			`+period+` AS period,
			COUNT(DISTINCT CASE
				WHEN user_id=0 THEN anonymous_user_id
				ELSE CAST(user_id AS TEXT)
				END) AS count
		FROM event_logs `+cond+`
		GROUP BY 1
	)
SELECT
	all_periods.period,
	COALESCE(count, 0)
FROM all_periods
LEFT OUTER JOIN users_by_period
ON all_periods.period = users_by_period.period
ORDER BY 1 DESC`, args...)
	if err != nil {
		return nil, err
	}

	counts := []UsageDatum{}
	defer rows.Close()
	for rows.Next() {
		var (
			c *int
			d *time.Time
		)
		err := rows.Scan(&d, &c)
		if err != nil {
			return nil, err
		}
		u := &UsageDatum{Count: *c, Start: (*d).UTC()}
		counts = append(counts, *u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

func (l *eventLogs) CountUniquesAll(ctx context.Context, startDate time.Time, endDate time.Time) (int, error) {
	return l.countUniquesBySQL(ctx, "WHERE DATE(timestamp) >= $1 AND DATE(timestamp) <= $2", startDate, endDate)
}

func (l *eventLogs) CountUniquesByEventNamePrefix(ctx context.Context, startDate time.Time, endDate time.Time, namePrefix string) (int, error) {
	return l.countUniquesBySQL(ctx, "WHERE DATE(timestamp) >= $1 AND DATE(timestamp) <= $2 AND name LIKE $3 || '%'", startDate, endDate, namePrefix)
}

func (l *eventLogs) CountUniquesByEventName(ctx context.Context, startDate time.Time, endDate time.Time, name string) (int, error) {
	return l.countUniquesBySQL(ctx, "WHERE DATE(timestamp) >= $1 AND DATE(timestamp) <= $2 AND name = $3", startDate, endDate, name)
}

func (l *eventLogs) CountUniquesByEventNames(ctx context.Context, startDate time.Time, endDate time.Time, names []string) (int, error) {
	items := []*sqlf.Query{}
	for _, v := range names {
		items = append(items, sqlf.Sprintf("%s", v))
	}
	q := sqlf.Sprintf("WHERE DATE(timestamp) >= %s AND DATE(timestamp) <= %s AND name IN (%s)", startDate.Format("2006-01-02 15:04:05 UTC"), endDate.Format("2006-01-02 15:04:05 UTC"), sqlf.Join(items, ","))
	return l.countUniquesBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (*eventLogs) countUniquesBySQL(ctx context.Context, query string, args ...interface{}) (int, error) {
	r := dbconn.Global.QueryRowContext(ctx, "SELECT COUNT(DISTINCT CASE WHEN user_id=0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) FROM event_logs "+query, args...)
	var count int
	err := r.Scan(&count)
	return count, err
}

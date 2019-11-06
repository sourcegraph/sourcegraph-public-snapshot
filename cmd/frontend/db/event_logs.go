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
}

func (*eventLogs) Insert(ctx context.Context, e *Event) error {
	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id, source, argument, version) VALUES($1, $2, $3, $4, $5, $6, $7)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Source,
		e.Argument,
		version.Version(),
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
	dates AS (SELECT generate_series(($1)::timestamp, ($2)::timestamp, interval '1 `+interval+`') AS period),
	users AS (SELECT `+period+` AS period, COUNT(DISTINCT CASE WHEN user_id=0 THEN anonymous_user_id ELSE CAST(user_id AS TEXT) END) AS count from event_logs `+cond+` GROUP BY 1)
SELECT
	dates.period,
	COALESCE(count, 0)
FROM dates LEFT OUTER JOIN users ON dates.period = users.period ORDER BY 1 DESC`, args...)
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
		u := &UsageDatum{Count: *c, Start: *d}
		counts = append(counts, *u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

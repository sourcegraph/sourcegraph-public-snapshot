package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Monitor struct {
	ID              int64
	CreatedBy       int32
	CreatedAt       time.Time
	ChangedBy       int32
	ChangedAt       time.Time
	Description     string
	Enabled         bool
	NamespaceUserID int32
}

// monitorColumns are the columns needed to fill out a Monitor.
// Its fields and order must be kept in sync with scanMonitor.
var monitorColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_monitors.id"),
	sqlf.Sprintf("cm_monitors.created_by"),
	sqlf.Sprintf("cm_monitors.created_at"),
	sqlf.Sprintf("cm_monitors.changed_by"),
	sqlf.Sprintf("cm_monitors.changed_at"),
	sqlf.Sprintf("cm_monitors.description"),
	sqlf.Sprintf("cm_monitors.enabled"),
	sqlf.Sprintf("cm_monitors.namespace_user_id"),
}

const insertCodeMonitorFmtStr = `
INSERT INTO cm_monitors
(created_at, created_by, changed_at, changed_by, description, enabled, namespace_user_id, namespace_org_id)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateMonitor(ctx context.Context, args *graphqlbackend.CreateMonitorArgs) (*Monitor, error) {
	var orgID, userID int32
	err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}

	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		insertCodeMonitorFmtStr,
		now,
		a.UID,
		now,
		a.UID,
		args.Description,
		args.Enabled,
		nilOrInt32(userID),
		nilOrInt32(orgID),
		sqlf.Join(monitorColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
}

func (s *codeMonitorStore) UpdateMonitor(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (*Monitor, error) {
	q, err := s.updateCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
}

func (s *codeMonitorStore) ToggleMonitor(ctx context.Context, args *graphqlbackend.ToggleCodeMonitorArgs) (*Monitor, error) {
	q, err := s.toggleCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
}

func (s *codeMonitorStore) DeleteMonitor(ctx context.Context, args *graphqlbackend.DeleteCodeMonitorArgs) error {
	q, err := s.deleteMonitorQuery(ctx, args)
	if err != nil {
		return err
	}
	return s.Exec(ctx, q)
}

func (s *codeMonitorStore) Monitors(ctx context.Context, userID int32, args *graphqlbackend.ListMonitorsArgs) ([]*Monitor, error) {
	q, err := monitorsQuery(userID, args)
	if err != nil {
		return nil, err
	}
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMonitors(rows)
}

const monitorByIDFmtStr = `
SELECT %s -- monitorColumns
FROM cm_monitors
WHERE id = %s
`

// GetMonitor fetches the monitor with the given ID, or returns sql.ErrNoRows if it does not exist
func (s *codeMonitorStore) GetMonitor(ctx context.Context, monitorID int64) (*Monitor, error) {
	q := sqlf.Sprintf(
		monitorByIDFmtStr,
		sqlf.Join(monitorColumns, ","),
		monitorID,
	)
	row := s.QueryRow(ctx, q)
	return scanMonitor(row)
}

const totalCountMonitorsFmtStr = `
SELECT COUNT(*)
FROM cm_monitors
WHERE namespace_user_id = %s;
`

func (s *codeMonitorStore) CountMonitors(ctx context.Context, userID int32) (int32, error) {
	var count int32
	err := s.QueryRow(ctx, sqlf.Sprintf(totalCountMonitorsFmtStr, userID)).Scan(&count)
	return count, err
}

const monitorsFmtStr = `
SELECT id, created_by, created_at, changed_by, changed_at, description, enabled, namespace_user_id
FROM cm_monitors
WHERE namespace_user_id = %s
AND id > %s
ORDER BY id ASC
LIMIT %s
`

func monitorsQuery(userID int32, args *graphqlbackend.ListMonitorsArgs) (*sqlf.Query, error) {
	after, err := unmarshalAfter(args.After)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		monitorsFmtStr,
		userID,
		after,
		args.First,
	), nil
}

const toggleCodeMonitorFmtStr = `
UPDATE cm_monitors
SET enabled = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
RETURNING %s
`

func (s *codeMonitorStore) toggleCodeMonitorQuery(ctx context.Context, args *graphqlbackend.ToggleCodeMonitorArgs) (*sqlf.Query, error) {
	var monitorID int64
	err := relay.UnmarshalSpec(args.Id, &monitorID)
	if err != nil {
		return nil, err
	}
	actorUID := actor.FromContext(ctx).UID
	return sqlf.Sprintf(
		toggleCodeMonitorFmtStr,
		args.Enabled,
		actorUID,
		s.Now(),
		monitorID,
		sqlf.Join(monitorColumns, ", "),
	), nil
}

const updateCodeMonitorFmtStr = `
UPDATE cm_monitors
SET description = %s,
	enabled = %s,
	namespace_user_id = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
RETURNING %s;
`

func (s *codeMonitorStore) updateCodeMonitorQuery(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (*sqlf.Query, error) {
	var userID, orgID int32
	err := graphqlbackend.UnmarshalNamespaceID(args.Monitor.Update.Namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}
	now := s.Now()
	a := actor.FromContext(ctx)
	var monitorID int64
	err = relay.UnmarshalSpec(args.Monitor.Id, &monitorID)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		updateCodeMonitorFmtStr,
		args.Monitor.Update.Description,
		args.Monitor.Update.Enabled,
		nilOrInt32(userID),
		a.UID,
		now,
		monitorID,
		sqlf.Join(monitorColumns, ", "),
	), nil
}

const deleteMonitorFmtStr = `
DELETE FROM cm_monitors
WHERE id = %s
`

func (s *codeMonitorStore) deleteMonitorQuery(ctx context.Context, args *graphqlbackend.DeleteCodeMonitorArgs) (*sqlf.Query, error) {
	var monitorID int64
	err := relay.UnmarshalSpec(args.Id, &monitorID)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		deleteMonitorFmtStr,
		monitorID,
	), nil
}

func scanMonitors(rows *sql.Rows) ([]*Monitor, error) {
	var ms []*Monitor
	for rows.Next() {
		m, err := scanMonitor(rows)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, rows.Err()
}

// scanMonitor scans a monitor from either a *sql.Row or *sql.Rows.
// It must be kept in sync with monitorColumns.
func scanMonitor(scanner dbutil.Scanner) (*Monitor, error) {
	m := &Monitor{}
	err := scanner.Scan(
		&m.ID,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.ChangedBy,
		&m.ChangedAt,
		&m.Description,
		&m.Enabled,
		&m.NamespaceUserID,
	)
	return m, err
}

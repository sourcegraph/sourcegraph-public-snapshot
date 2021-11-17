package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
)

type MonitorEmail struct {
	Id        int64
	Monitor   int64
	Enabled   bool
	Priority  string
	Header    string
	CreatedBy int32
	CreatedAt time.Time
	ChangedBy int32
	ChangedAt time.Time
}

func (s *codeMonitorStore) UpdateActionEmail(ctx context.Context, monitorID int64, action *graphqlbackend.EditActionArgs) (e *MonitorEmail, err error) {
	var q *sqlf.Query
	q, err = s.updateActionEmailQuery(ctx, monitorID, action.Email)
	if err != nil {
		return nil, err
	}
	e, err = s.runEmailQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *codeMonitorStore) CreateActionEmail(ctx context.Context, monitorID int64, action *graphqlbackend.CreateActionArgs) (e *MonitorEmail, err error) {
	var q *sqlf.Query
	q, err = s.createActionEmailQuery(ctx, monitorID, action.Email)
	if err != nil {
		return nil, err
	}
	e, err = s.runEmailQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *codeMonitorStore) DeleteActionsInt64(ctx context.Context, actionIDs []int64, monitorID int64) (err error) {
	if len(actionIDs) == 0 {
		return nil
	}
	var q *sqlf.Query
	q, err = deleteActionsEmailQuery(ctx, actionIDs, monitorID)
	if err != nil {
		return err
	}
	err = s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

const totalCountActionEmailsFmtStr = `
SELECT COUNT(*)
FROM cm_emails
WHERE monitor = %s;
`

func (s *codeMonitorStore) TotalCountActionEmails(ctx context.Context, monitorID int64) (count int32, err error) {
	err = s.QueryRow(ctx, sqlf.Sprintf(totalCountActionEmailsFmtStr, monitorID)).Scan(&count)
	return count, err
}

const actionEmailByIDFmtStr = `
SELECT id, monitor, enabled, priority, header, created_by, created_at, changed_by, changed_at
FROM cm_emails
WHERE id = %s
`

func (s *codeMonitorStore) ActionEmailByIDInt64(ctx context.Context, emailID int64) (m *MonitorEmail, err error) {
	return s.runEmailQuery(ctx, sqlf.Sprintf(actionEmailByIDFmtStr, emailID))
}

func (s *codeMonitorStore) runEmailQuery(ctx context.Context, q *sqlf.Query) (*MonitorEmail, error) {
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	es, err := scanEmails(rows)
	if err != nil {
		return nil, err
	}
	if len(es) == 0 {
		return nil, errors.Errorf("operation failed. Query should have returned 1 row")
	}
	return es[0], nil
}

const updateActionEmailFmtStr = `
UPDATE cm_emails
SET enabled = %s,
	priority = %s,
	header = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
AND monitor = %s
RETURNING %s;
`

func (s *codeMonitorStore) updateActionEmailQuery(ctx context.Context, monitorID int64, args *graphqlbackend.EditActionEmailArgs) (q *sqlf.Query, err error) {
	var actionID int64
	if args.Id == nil {
		return nil, errors.Errorf("nil is not a valid action ID")
	}
	err = relay.UnmarshalSpec(*args.Id, &actionID)
	if err != nil {
		return nil, err
	}
	now := s.Now()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		updateActionEmailFmtStr,
		args.Update.Enabled,
		args.Update.Priority,
		args.Update.Header,
		a.UID,
		now,
		actionID,
		monitorID,
		sqlf.Join(EmailsColumns, ", "),
	), nil
}

// ListActionsOpts holds list options for listing actions
type ListActionsOpts struct {
	// MonitorID, if set, will constrain the listed actions to only
	// those that are defined as part of the given monitor.
	// References cm_monitors(id)
	MonitorID *int

	// First, if set, limits the number of actions returned
	// to the first n.
	First *int

	// After, if set, begins listing actions after the given id
	After *int
}

func (o ListActionsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.MonitorID != nil {
		conds = append(conds, sqlf.Sprintf("monitor = %s", *o.MonitorID))
	}
	if o.After != nil {
		conds = append(conds, sqlf.Sprintf("id > %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

func (o ListActionsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const listEmailActionsFmtStr = `
SELECT id, monitor, enabled, priority, header, created_by, created_at, changed_by, changed_at
FROM cm_emails
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

// ListEmailActions lists emails from cm_emails with the given opts
func (s *codeMonitorStore) ListEmailActions(ctx context.Context, opts ListActionsOpts) ([]*MonitorEmail, error) {
	q := sqlf.Sprintf(
		listEmailActionsFmtStr,
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmails(rows)
}

const createActionEmailFmtStr = `
INSERT INTO cm_emails
(monitor, enabled, priority, header, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) createActionEmailQuery(ctx context.Context, monitorID int64, args *graphqlbackend.CreateActionEmailArgs) (*sqlf.Query, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		createActionEmailFmtStr,
		monitorID,
		args.Enabled,
		args.Priority,
		args.Header,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(EmailsColumns, ", "),
	), nil
}

const deleteActionEmailFmtStr = `DELETE FROM cm_emails WHERE id in (%s) AND MONITOR = %s`

func deleteActionsEmailQuery(ctx context.Context, actionIDs []int64, monitorID int64) (*sqlf.Query, error) {
	var deleteIDs []*sqlf.Query
	for _, ids := range actionIDs {
		deleteIDs = append(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	return sqlf.Sprintf(
		deleteActionEmailFmtStr,
		sqlf.Join(deleteIDs, ", "),
		monitorID,
	), nil
}

var EmailsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_emails.id"),
	sqlf.Sprintf("cm_emails.monitor"),
	sqlf.Sprintf("cm_emails.enabled"),
	sqlf.Sprintf("cm_emails.priority"),
	sqlf.Sprintf("cm_emails.header"),
	sqlf.Sprintf("cm_emails.created_by"),
	sqlf.Sprintf("cm_emails.created_at"),
	sqlf.Sprintf("cm_emails.changed_by"),
	sqlf.Sprintf("cm_emails.changed_at"),
}

func scanEmails(rows *sql.Rows) (ms []*MonitorEmail, err error) {
	for rows.Next() {
		m := &MonitorEmail{}
		if err = rows.Scan(
			&m.Id,
			&m.Monitor,
			&m.Enabled,
			&m.Priority,
			&m.Header,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.ChangedBy,
			&m.ChangedAt,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

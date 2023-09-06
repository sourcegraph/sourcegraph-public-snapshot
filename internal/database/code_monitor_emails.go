package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type EmailAction struct {
	ID             int64
	Monitor        int64
	Enabled        bool
	Priority       string
	Header         string
	IncludeResults bool
	CreatedBy      int32
	CreatedAt      time.Time
	ChangedBy      int32
	ChangedAt      time.Time
}

const updateActionEmailFmtStr = `
UPDATE cm_emails
SET enabled = %s,
    include_results = %s,
	priority = %s,
	header = %s,
	changed_by = %s,
	changed_at = %s
WHERE
	id = %s
	AND EXISTS (
		SELECT 1 FROM cm_monitors
		WHERE cm_monitors.id = cm_emails.monitor
			AND %s
	)
RETURNING %s;
`

type EmailActionArgs struct {
	Enabled        bool
	IncludeResults bool
	Priority       string
	Header         string
}

func (s *codeMonitorStore) UpdateEmailAction(ctx context.Context, id int64, args *EmailActionArgs) (*EmailAction, error) {
	a := actor.FromContext(ctx)

	user, err := a.User(ctx, s.userStore)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		updateActionEmailFmtStr,
		args.Enabled,
		args.IncludeResults,
		args.Priority,
		args.Header,
		a.UID,
		s.Now(),
		id,
		namespaceScopeQuery(user),
		sqlf.Join(emailsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scanEmail(row)
}

const createActionEmailFmtStr = `
INSERT INTO cm_emails
(monitor, enabled, include_results, priority, header, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateEmailAction(ctx context.Context, monitorID int64, args *EmailActionArgs) (*EmailAction, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		createActionEmailFmtStr,
		monitorID,
		args.Enabled,
		args.IncludeResults,
		args.Priority,
		args.Header,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(emailsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	return scanEmail(row)
}

const deleteActionEmailFmtStr = `
DELETE FROM cm_emails
WHERE id in (%s)
	AND MONITOR = %s
`

func (s *codeMonitorStore) DeleteEmailActions(ctx context.Context, actionIDs []int64, monitorID int64) error {
	if len(actionIDs) == 0 {
		return nil
	}

	deleteIDs := make([]*sqlf.Query, 0, len(actionIDs))
	for _, ids := range actionIDs {
		deleteIDs = append(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	q := sqlf.Sprintf(
		deleteActionEmailFmtStr,
		sqlf.Join(deleteIDs, ", "),
		monitorID,
	)

	return s.Exec(ctx, q)
}

const actionEmailByIDFmtStr = `
SELECT %s -- EmailsColumns
FROM cm_emails
WHERE id = %s
`

func (s *codeMonitorStore) GetEmailAction(ctx context.Context, emailID int64) (m *EmailAction, err error) {
	q := sqlf.Sprintf(
		actionEmailByIDFmtStr,
		sqlf.Join(emailsColumns, ","),
		emailID,
	)
	row := s.QueryRow(ctx, q)
	return scanEmail(row)
}

// ListActionsOpts holds list options for listing actions
type ListActionsOpts struct {
	// MonitorID, if set, will constrain the listed actions to only
	// those that are defined as part of the given monitor.
	// References cm_monitors(id)
	MonitorID *int64

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
SELECT %s -- EmailsColumns
FROM cm_emails
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

// ListEmailActions lists emails from cm_emails with the given opts
func (s *codeMonitorStore) ListEmailActions(ctx context.Context, opts ListActionsOpts) ([]*EmailAction, error) {
	q := sqlf.Sprintf(
		listEmailActionsFmtStr,
		sqlf.Join(emailsColumns, ","),
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

// emailColumns is the set of columns in the cm_emails table
// This must be kept in sync with scanEmail
var emailsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_emails.id"),
	sqlf.Sprintf("cm_emails.monitor"),
	sqlf.Sprintf("cm_emails.enabled"),
	sqlf.Sprintf("cm_emails.priority"),
	sqlf.Sprintf("cm_emails.header"),
	sqlf.Sprintf("cm_emails.include_results"),
	sqlf.Sprintf("cm_emails.created_by"),
	sqlf.Sprintf("cm_emails.created_at"),
	sqlf.Sprintf("cm_emails.changed_by"),
	sqlf.Sprintf("cm_emails.changed_at"),
}

func scanEmails(rows *sql.Rows) ([]*EmailAction, error) {
	var ms []*EmailAction
	for rows.Next() {
		m, err := scanEmail(rows)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, rows.Err()
}

// scanEmail scans a MonitorEmail from a *sql.Row or *sql.Rows.
// It must be kept in sync with emailsColumns.
func scanEmail(scanner dbutil.Scanner) (*EmailAction, error) {
	m := &EmailAction{}
	err := scanner.Scan(
		&m.ID,
		&m.Monitor,
		&m.Enabled,
		&m.Priority,
		&m.Header,
		&m.IncludeResults,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.ChangedBy,
		&m.ChangedAt,
	)
	return m, err
}

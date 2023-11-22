package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type SlackWebhookAction struct {
	ID             int64
	Monitor        int64
	Enabled        bool
	URL            string
	IncludeResults bool

	CreatedBy int32
	CreatedAt time.Time
	ChangedBy int32
	ChangedAt time.Time
}

const updateSlackWebhookActionQuery = `
UPDATE cm_slack_webhooks
SET enabled = %s,
	include_results = %s,
	url = %s,
	changed_by = %s,
	changed_at = %s
WHERE
	id = %s
	AND EXISTS (
		SELECT 1 FROM cm_monitors
		WHERE cm_monitors.id = cm_slack_webhooks.monitor
			AND %s
	)
RETURNING %s;
`

func (s *codeMonitorStore) UpdateSlackWebhookAction(ctx context.Context, id int64, enabled, includeResults bool, url string) (*SlackWebhookAction, error) {
	a := actor.FromContext(ctx)

	user, err := a.User(ctx, s.userStore)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		updateSlackWebhookActionQuery,
		enabled,
		includeResults,
		url,
		a.UID,
		s.Now(),
		id,
		namespaceScopeQuery(user),
		sqlf.Join(slackWebhookActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanSlackWebhookAction(row)
}

const createSlackWebhookActionQuery = `
INSERT INTO cm_slack_webhooks
(monitor, enabled, include_results, url, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) CreateSlackWebhookAction(ctx context.Context, monitorID int64, enabled, includeResults bool, url string) (*SlackWebhookAction, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	q := sqlf.Sprintf(
		createSlackWebhookActionQuery,
		monitorID,
		enabled,
		includeResults,
		url,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(slackWebhookActionColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanSlackWebhookAction(row)
}

const deleteSlackWebhookActionQuery = `
DELETE FROM cm_slack_webhooks
WHERE id in (%s)
	AND MONITOR = %s
`

func (s *codeMonitorStore) DeleteSlackWebhookActions(ctx context.Context, monitorID int64, webhookIDs ...int64) error {
	if len(webhookIDs) == 0 {
		return nil
	}

	deleteIDs := make([]*sqlf.Query, 0, len(webhookIDs))
	for _, ids := range webhookIDs {
		deleteIDs = append(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	q := sqlf.Sprintf(
		deleteSlackWebhookActionQuery,
		sqlf.Join(deleteIDs, ","),
		monitorID,
	)

	return s.Exec(ctx, q)
}

const countSlackWebhookActionsQuery = `
SELECT COUNT(*)
FROM cm_slack_webhooks
WHERE monitor = %s;
`

func (s *codeMonitorStore) CountSlackWebhookActions(ctx context.Context, monitorID int64) (int, error) {
	var count int
	err := s.QueryRow(ctx, sqlf.Sprintf(countSlackWebhookActionsQuery, monitorID)).Scan(&count)
	return count, err
}

const getSlackWebhookActionQuery = `
SELECT %s -- SlackWebhookActionColumns
FROM cm_slack_webhooks
WHERE id = %s
`

func (s *codeMonitorStore) GetSlackWebhookAction(ctx context.Context, id int64) (*SlackWebhookAction, error) {
	q := sqlf.Sprintf(
		getSlackWebhookActionQuery,
		sqlf.Join(slackWebhookActionColumns, ","),
		id,
	)
	row := s.QueryRow(ctx, q)
	return scanSlackWebhookAction(row)
}

const listSlackWebhookActionsQuery = `
SELECT %s -- SlackWebhookActionColumns
FROM cm_slack_webhooks
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

func (s *codeMonitorStore) ListSlackWebhookActions(ctx context.Context, opts ListActionsOpts) ([]*SlackWebhookAction, error) {
	q := sqlf.Sprintf(
		listSlackWebhookActionsQuery,
		sqlf.Join(slackWebhookActionColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSlackWebhookActions(rows)
}

// slackWebhookActionColumns is the set of columns in the cm_slack_webhooks table
// This must be kept in sync with scanSlackWebhook
var slackWebhookActionColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_slack_webhooks.id"),
	sqlf.Sprintf("cm_slack_webhooks.monitor"),
	sqlf.Sprintf("cm_slack_webhooks.enabled"),
	sqlf.Sprintf("cm_slack_webhooks.url"),
	sqlf.Sprintf("cm_slack_webhooks.include_results"),
	sqlf.Sprintf("cm_slack_webhooks.created_by"),
	sqlf.Sprintf("cm_slack_webhooks.created_at"),
	sqlf.Sprintf("cm_slack_webhooks.changed_by"),
	sqlf.Sprintf("cm_slack_webhooks.changed_at"),
}

func scanSlackWebhookActions(rows *sql.Rows) ([]*SlackWebhookAction, error) {
	var ws []*SlackWebhookAction
	for rows.Next() {
		w, err := scanSlackWebhookAction(rows)
		if err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, rows.Err()
}

// scanSlackWebhookAction scans a SlackWebhookAction from a *sql.Row or *sql.Rows.
// It must be kept in sync with slackWebhookActionColumns.
func scanSlackWebhookAction(scanner dbutil.Scanner) (*SlackWebhookAction, error) {
	var w SlackWebhookAction
	err := scanner.Scan(
		&w.ID,
		&w.Monitor,
		&w.Enabled,
		&w.URL,
		&w.IncludeResults,
		&w.CreatedBy,
		&w.CreatedAt,
		&w.ChangedBy,
		&w.ChangedAt,
	)
	return &w, err
}

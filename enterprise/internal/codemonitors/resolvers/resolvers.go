package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// NewResolver returns a new Resolver that uses the given db
func NewResolver(db dbutil.DB) graphqlbackend.CodeMonitorsResolver {
	return &Resolver{db: basestore.NewWithDB(db, sql.TxOptions{}), clock: func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }}
}

// newResolverWithClock is used in tests to set the clock manually.
func newResolverWithClock(db dbutil.DB, clock func() time.Time) graphqlbackend.CodeMonitorsResolver {
	return &Resolver{db: basestore.NewWithDB(db, sql.TxOptions{}), clock: clock}
}

type Resolver struct {
	db    *basestore.Store
	clock func() time.Time
}

func (r *Resolver) Monitors(ctx context.Context, userID int32, args *graphqlbackend.ListMonitorsArgs) (graphqlbackend.MonitorConnectionResolver, error) {
	q, err := monitorsQuery(userID, args)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ms, err := scanMonitors(rows)
	if err != nil {
		return nil, err
	}
	// Hydrate the monitors with the resolver.
	for _, m := range ms {
		m.(*monitor).Resolver = r
	}
	return &monitorConnection{r, ms}, nil
}

func (r *Resolver) CreateCodeMonitor(ctx context.Context, args *graphqlbackend.CreateCodeMonitorArgs) (m graphqlbackend.MonitorResolver, err error) {
	// TODO: check actor is the owner, or site-admin, or part of the owner-org
	// start transaction
	var txStore *basestore.Store
	txStore, err = r.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	tx := Resolver{
		db:    txStore,
		clock: r.clock,
	}
	defer func() { err = tx.db.Done(err) }()

	// create code monitor
	var q *sqlf.Query
	q, err = tx.createCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	m, err = tx.runMonitorQuery(ctx, q)
	if err != nil {
		return nil, err
	}

	// create trigger
	q, err = tx.createTriggerQueryQuery(ctx, m.(*monitor).id, args)
	if err != nil {
		return nil, err
	}
	err = tx.db.Exec(ctx, q)
	if err != nil {
		return nil, err
	}

	// create actions
	for i, action := range args.Actions {
		if action.Email != nil {
			q, err = tx.createActionEmailQuery(ctx, m.(*monitor).id, action.Email)
			if err != nil {
				return nil, err
			}
			var e graphqlbackend.MonitorEmailResolver
			e, err = tx.runEmailQuery(ctx, q)
			if err != nil {
				return nil, err
			}

			// insert recipients
			for _, recipient := range action.Email.Recipients {
				q, err = tx.createRecipientQuery(ctx, e.(*monitorEmail).id, recipient)
				if err != nil {
					return nil, err
				}
				err = tx.db.Exec(ctx, q)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return nil, fmt.Errorf("missing email object for action %d", i)
		}
	}
	// Hydrate monitor with Resolver.
	m.(*monitor).Resolver = r
	return m, nil
}

func (r *Resolver) ToggleCodeMonitor(ctx context.Context, args *graphqlbackend.ToggleCodeMonitorArgs) (graphqlbackend.MonitorResolver, error) {
	err := r.isAllowedToEdit(ctx, args.Id)
	if err != nil {
		return nil, fmt.Errorf("ToggleCodeMonitor: %w", err)
	}
	q, err := r.toggleCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	return r.runMonitorQuery(ctx, q)
}

func (r *Resolver) DeleteCodeMonitor(ctx context.Context, args *graphqlbackend.DeleteCodeMonitorArgs) (*graphqlbackend.EmptyResponse, error) {
	err := r.isAllowedToEdit(ctx, args.Id)
	if err != nil {
		return nil, fmt.Errorf("DeleteCodeMonitor: %w", err)
	}
	q, err := r.deleteCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	if err := r.db.Exec(ctx, q); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) UpdateCodeMonitor(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (m graphqlbackend.MonitorResolver, err error) {
	err = r.isAllowedToEdit(ctx, args.Monitor.Id)
	if err != nil {
		return nil, fmt.Errorf("UpdateCodeMonitor: %w", err)
	}
	var (
		q  *sqlf.Query
		tx *Resolver
	)
	tx, err = r.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.db.Done(err) }()

	// Update monitor.
	q, err = tx.updateCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	m, err = tx.runMonitorQuery(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("update monitor: %w", err)
	}

	// Update trigger.
	q, err = tx.updateTriggerQueryQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	err = tx.db.Exec(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("update trigger: %w", err)
	}

	// Update actions.
	q, err = tx.deleteRecipientQuery(ctx, m.(*monitor).id)
	if err != nil {
		return nil, err
	}
	err = tx.db.Exec(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("delete recipient: %w", err)
	}
	for i, action := range args.Actions {
		if action.Email != nil {
			q, err = tx.updateActionEmailQuery(ctx, m.(*monitor).id, action.Email)
			if err != nil {
				return nil, err
			}
			var e graphqlbackend.MonitorEmailResolver
			e, err = tx.runEmailQuery(ctx, q)
			if err != nil {
				return nil, fmt.Errorf("update email: %w", err)
			}

			// Insert recipients.
			for _, recipient := range action.Email.Update.Recipients {
				q, err = tx.createRecipientQuery(ctx, e.(*monitorEmail).id, recipient)
				if err != nil {
					return nil, err
				}
				err = tx.db.Exec(ctx, q)
				if err != nil {
					return nil, fmt.Errorf("create recipient: %w", err)
				}
			}
		} else {
			return nil, fmt.Errorf("missing email object for action %d", i)
		}
	}
	// Hydrate monitor with Resolver.
	m.(*monitor).Resolver = r
	return m, nil
}

func (r *Resolver) transact(ctx context.Context) (*Resolver, error) {
	txStore, err := r.db.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Resolver{
		db:    txStore,
		clock: r.clock,
	}, nil
}

func (r *Resolver) runMonitorQuery(ctx context.Context, q *sqlf.Query) (graphqlbackend.MonitorResolver, error) {
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ms, err := scanMonitors(rows)
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, fmt.Errorf("operation failed. Query should have returned 1 row")
	}
	return ms[0], nil
}

func (r *Resolver) runEmailQuery(ctx context.Context, q *sqlf.Query) (graphqlbackend.MonitorEmailResolver, error) {
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ms, err := scanEmails(rows)
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, fmt.Errorf("operation failed. Query should have returned 1 row")
	}
	return ms[0], nil
}

func (r *Resolver) runTriggerQuery(ctx context.Context, q *sqlf.Query) (graphqlbackend.MonitorQueryResolver, error) {
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ms, err := scanQueries(rows)
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, fmt.Errorf("operation failed. Query should have returned 1 row")
	}
	return ms[0], nil
}

func scanMonitors(rows *sql.Rows) ([]graphqlbackend.MonitorResolver, error) {
	var ms []graphqlbackend.MonitorResolver
	for rows.Next() {
		m := &monitor{}
		if err := rows.Scan(
			&m.id,
			&m.createdBy,
			&m.createdAt,
			&m.changedBy,
			&m.changedAt,
			&m.description,
			&m.enabled,
			&m.namespaceUserID,
			&m.namespaceOrgID,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err := rows.Close()
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

func scanEmails(rows *sql.Rows) ([]graphqlbackend.MonitorEmailResolver, error) {
	var ms []graphqlbackend.MonitorEmailResolver
	for rows.Next() {
		m := &monitorEmail{}
		if err := rows.Scan(
			&m.id,
			&m.monitor,
			&m.enabled,
			&m.priority,
			&m.header,
			&m.createdBy,
			&m.createdAt,
			&m.changedBy,
			&m.changedAt,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err := rows.Close()
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

func scanQueries(rows *sql.Rows) ([]graphqlbackend.MonitorQueryResolver, error) {
	var ms []graphqlbackend.MonitorQueryResolver
	for rows.Next() {
		m := &monitorQuery{}
		if err := rows.Scan(
			&m.id,
			&m.monitor,
			&m.query,
			&m.createdBy,
			&m.createdAt,
			&m.changedBy,
			&m.changedAt,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err := rows.Close()
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

type recipient struct {
	id              int64
	email           int64
	namespaceUserID *int32
	namespaceOrgID  *int32
}

func scanRecipients(rows *sql.Rows) (ms []*recipient, err error) {
	for rows.Next() {
		m := &recipient{}
		if err := rows.Scan(
			&m.id,
			&m.email,
			&m.namespaceUserID,
			&m.namespaceOrgID,
		); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return ms, nil
}

func nilOrInt32(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

var monitorColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_monitors.id"),
	sqlf.Sprintf("cm_monitors.created_by"),
	sqlf.Sprintf("cm_monitors.created_at"),
	sqlf.Sprintf("cm_monitors.changed_by"),
	sqlf.Sprintf("cm_monitors.changed_at"),
	sqlf.Sprintf("cm_monitors.description"),
	sqlf.Sprintf("cm_monitors.enabled"),
	sqlf.Sprintf("cm_monitors.namespace_user_id"),
	sqlf.Sprintf("cm_monitors.namespace_org_id"),
}

func monitorsQuery(userID int32, args *graphqlbackend.ListMonitorsArgs) (*sqlf.Query, error) {
	const SelectMonitorsByOwner = `
SELECT id, created_by, created_at, changed_by, changed_at, description, enabled, namespace_user_id, namespace_org_id 
FROM cm_monitors
WHERE namespace_user_id = %s
AND id > %s
ORDER BY id ASC
LIMIT %s
`
	after, err := unmarshallAfter(args.After)
	if err != nil {
		return nil, err
	}
	query := sqlf.Sprintf(
		SelectMonitorsByOwner,
		userID,
		after,
		args.First,
	)
	return query, nil
}

func (r *Resolver) createCodeMonitorQuery(ctx context.Context, args *graphqlbackend.CreateCodeMonitorArgs) (*sqlf.Query, error) {
	const InsertCodeMonitorQuery = `
INSERT INTO cm_monitors 
(created_at, created_by, changed_at, changed_by, description, enabled, namespace_user_id, namespace_org_id) 
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`
	var userID int32
	var orgID int32
	err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}
	now := r.clock()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		InsertCodeMonitorQuery,
		now,
		a.UID,
		now,
		a.UID,
		args.Description,
		args.Enabled,
		nilOrInt32(userID),
		nilOrInt32(orgID),
		sqlf.Join(monitorColumns, ", "),
	), nil
}

func (r *Resolver) updateCodeMonitorQuery(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (*sqlf.Query, error) {
	const updateCodeMonitorQuery = `
UPDATE cm_monitors
SET description = %s,
	enabled	= %s,
	namespace_user_id = %s,
	namespace_org_id = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
RETURNING %s;
`
	var userID int32
	var orgID int32
	err := graphqlbackend.UnmarshalNamespaceID(args.Monitor.Update.Namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}
	now := r.clock()
	a := actor.FromContext(ctx)
	var monitorID int64
	err = relay.UnmarshalSpec(args.Monitor.Id, &monitorID)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		updateCodeMonitorQuery,
		args.Monitor.Update.Description,
		args.Monitor.Update.Enabled,
		nilOrInt32(userID),
		nilOrInt32(orgID),
		a.UID,
		now,
		monitorID,
		sqlf.Join(monitorColumns, ", "),
	), nil
}

var queryColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_queries.id"),
	sqlf.Sprintf("cm_queries.monitor"),
	sqlf.Sprintf("cm_queries.query"),
	sqlf.Sprintf("cm_queries.created_by"),
	sqlf.Sprintf("cm_queries.created_at"),
	sqlf.Sprintf("cm_queries.changed_by"),
	sqlf.Sprintf("cm_queries.changed_at"),
}

func (r *Resolver) createTriggerQueryQuery(ctx context.Context, monitorId int64, args *graphqlbackend.CreateCodeMonitorArgs) (*sqlf.Query, error) {
	const insertQueryQuery = `
INSERT INTO cm_queries
(monitor, query, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s)
RETURNING %s;
`
	var userID int32
	var orgID int32
	err := graphqlbackend.UnmarshalNamespaceID(args.Namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}
	now := r.clock()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		insertQueryQuery,
		monitorId,
		args.Trigger.Query,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(queryColumns, ", "),
	), nil
}

func (r *Resolver) updateTriggerQueryQuery(ctx context.Context, args *graphqlbackend.UpdateCodeMonitorArgs) (q *sqlf.Query, err error) {
	const updateTriggerQueryQuery = `
UPDATE cm_queries
SET query = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
AND monitor = %s
RETURNING %s;
`
	now := r.clock()
	a := actor.FromContext(ctx)

	var triggerID int64
	err = relay.UnmarshalSpec(args.Trigger.Id, &triggerID)
	if err != nil {
		return nil, err
	}

	var monitorID int64
	err = relay.UnmarshalSpec(args.Monitor.Id, &monitorID)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		updateTriggerQueryQuery,
		args.Trigger.Update.Query,
		a.UID,
		now,
		triggerID,
		monitorID,
		sqlf.Join(queryColumns, ", "),
	), nil
}

func (r *Resolver) triggerQueryByMonitorQuery(ctx context.Context, monitorID int64) (*sqlf.Query, error) {
	const triggerQueryByMonitorQuery = `
SELECT id, monitor, query, created_by, created_at, changed_by, changed_at
FROM cm_queries
WHERE monitor = %s;
`
	return sqlf.Sprintf(
		triggerQueryByMonitorQuery,
		monitorID,
	), nil
}

var emailsColumns = []*sqlf.Query{
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

func (r *Resolver) createActionEmailQuery(ctx context.Context, monitorId int64, args *graphqlbackend.CreateActionEmailArgs) (*sqlf.Query, error) {
	const insertEmailQuery = `
INSERT INTO cm_emails
(monitor, enabled, priority, header, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`
	now := r.clock()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		insertEmailQuery,
		monitorId,
		args.Enabled,
		args.Priority,
		args.Header,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(emailsColumns, ", "),
	), nil
}

var recipientsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_recipients.id"),
	sqlf.Sprintf("cm_recipients.email"),
	sqlf.Sprintf("cm_recipients.namespace_user_id"),
	sqlf.Sprintf("cm_recipients.namespace_org_id"),
}

func (r *Resolver) updateActionEmailQuery(ctx context.Context, monitorID int64, args *graphqlbackend.EditActionEmailArgs) (q *sqlf.Query, err error) {
	const updateMonitorActionEmailQuery = `
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
	var actionID int64
	err = relay.UnmarshalSpec(args.Id, &actionID)
	if err != nil {
		return nil, err
	}
	now := r.clock()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		updateMonitorActionEmailQuery,
		args.Update.Enabled,
		args.Update.Priority,
		args.Update.Header,
		a.UID,
		now,
		actionID,
		monitorID,
		sqlf.Join(emailsColumns, ", "),
	), nil
}

func (r *Resolver) readActionEmailQuery(ctx context.Context, monitorId int64, args *graphqlbackend.ListActionArgs) (*sqlf.Query, error) {
	const readActionEmailQuery = `
SELECT id, monitor, enabled, priority, header, created_by, created_at, changed_by, changed_at
FROM cm_emails
WHERE monitor = %s;
`
	return sqlf.Sprintf(
		readActionEmailQuery,
		monitorId,
	), nil
}

func (r *Resolver) createRecipientQuery(ctx context.Context, emailId int64, namespace graphql.ID) (*sqlf.Query, error) {
	const insertRecipientQuery = `
INSERT INTO cm_recipients
(email, namespace_user_id, namespace_org_id)
VALUES (%s,%s,%s)
RETURNING %s;
`
	var userID int32
	var orgID int32
	err := graphqlbackend.UnmarshalNamespaceID(namespace, &userID, &orgID)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		insertRecipientQuery,
		emailId,
		nilOrInt32(userID),
		nilOrInt32(orgID),
		sqlf.Join(recipientsColumns, ", "),
	), nil
}

func (r *Resolver) deleteRecipientQuery(ctx context.Context, emailId int64) (*sqlf.Query, error) {
	const deleteRecipientQuery = `DELETE FROM cm_recipients WHERE email = %s`
	return sqlf.Sprintf(
		deleteRecipientQuery,
		emailId,
	), nil
}

func (r *Resolver) readRecipientQuery(ctx context.Context, emailId int64, args *graphqlbackend.ListRecipientsArgs) (*sqlf.Query, error) {
	const readRecipientQuery = `
SELECT id, email, namespace_user_id, namespace_org_id
FROM cm_recipients
WHERE email = %s
AND id > %s
ORDER BY id ASC
LIMIT %s;
`
	after, err := unmarshallAfter(args.After)
	if err != nil {
		return nil, err
	}
	return sqlf.Sprintf(
		readRecipientQuery,
		emailId,
		after,
		args.First,
	), nil
}

func (r *Resolver) toggleCodeMonitorQuery(ctx context.Context, args *graphqlbackend.ToggleCodeMonitorArgs) (*sqlf.Query, error) {
	toggleCodeMonitorQuery := `
UPDATE cm_monitors
SET enabled = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
RETURNING %s
`
	var monitorID int64
	err := relay.UnmarshalSpec(args.Id, &monitorID)
	if err != nil {
		return nil, err
	}
	actorUID := actor.FromContext(ctx).UID
	query := sqlf.Sprintf(
		toggleCodeMonitorQuery,
		args.Enabled,
		actorUID,
		r.clock(),
		monitorID,
		sqlf.Join(monitorColumns, ", "),
	)
	return query, nil
}

func (r *Resolver) deleteCodeMonitorQuery(ctx context.Context, args *graphqlbackend.DeleteCodeMonitorArgs) (*sqlf.Query, error) {
	deleteCodeMonitorQuery := `DELETE FROM cm_monitors WHERE id = %s`
	var monitorID int64
	err := relay.UnmarshalSpec(args.Id, &monitorID)
	if err != nil {
		return nil, err
	}
	query := sqlf.Sprintf(
		deleteCodeMonitorQuery,
		monitorID,
	)
	return query, nil
}

// isAllowedToEdit compares the owner of a monitor (user or org) to the actor of
// the request. A user can edit a monitor if either of the following statements
// is true:
// - she is the owner
// - she is a member of the organization which is the owner of the monitor
// - she is a site-admin
func (r *Resolver) isAllowedToEdit(ctx context.Context, id graphql.ID) error {
	var monitorId int32
	err := relay.UnmarshalSpec(id, &monitorId)
	if err != nil {
		return err
	}
	userId, orgID, err := r.ownerForId32(ctx, monitorId)
	if err != nil {
		return err
	}
	if userId == nil && orgID == nil {
		return fmt.Errorf("monitor does not exist")
	}
	if orgID != nil {
		if err := backend.CheckOrgAccess(ctx, *orgID); err != nil {
			return fmt.Errorf("user is not a member of the organization which owns the code monitor")
		}
		return nil
	}
	if err := backend.CheckSiteAdminOrSameUser(ctx, *userId); err != nil {
		return fmt.Errorf("user not allowed to edit")
	}
	return nil
}

func (r *Resolver) ownerForId32(ctx context.Context, monitorId int32) (userId *int32, orgId *int32, err error) {
	var (
		q    *sqlf.Query
		rows *sql.Rows
	)
	q, err = ownerForId32Query(ctx, monitorId)
	if err != nil {
		return nil, nil, err
	}
	rows, err = r.db.Query(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(
			&userId,
			&orgId,
		); err != nil {
			return nil, nil, err
		}
	}
	err = rows.Close()
	if err != nil {
		return nil, nil, err
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return userId, orgId, nil
}

func ownerForId32Query(ctx context.Context, monitorId int32) (*sqlf.Query, error) {
	const ownerForId32Query = `SELECT namespace_user_id, namespace_org_id FROM cm_monitors WHERE id = %s`
	return sqlf.Sprintf(
		ownerForId32Query,
		monitorId,
	), nil
}

//
// MonitorConnection
//
type monitorConnection struct {
	*Resolver
	monitors []graphqlbackend.MonitorResolver
}

func (m *monitorConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorResolver, error) {
	return m.monitors, nil
}

func (m *monitorConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(m.monitors)), nil
}

func (m *monitorConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if len(m.monitors) == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(string(m.monitors[len(m.monitors)-1].ID())), nil
}

//
// Monitor
//
type monitor struct {
	*Resolver
	id              int64
	createdBy       int32
	createdAt       time.Time
	changedBy       int32
	changedAt       time.Time
	description     string
	enabled         bool
	namespaceUserID *int32
	namespaceOrgID  *int32
}

const (
	monitorKind                     = "CodeMonitor"
	monitorTriggerQueryKind         = "CodeMonitorTriggerQuery"
	monitorTriggerEventKind         = "CodeMonitorTriggerEvent"
	monitorActionEmailKind          = "CodeMonitorActionEmail"
	monitorActionEmailRecipientKind = "CodeMonitorActionEmailRecipient"
)

func (m *monitor) ID() graphql.ID {
	return relay.MarshalID(monitorKind, m.id)
}

func (m *monitor) CreatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, m.createdBy)
}

func (m *monitor) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: m.createdAt}
}

func (m *monitor) Description() string {
	return m.description
}

func (m *monitor) Enabled() bool {
	return m.enabled
}

func (m *monitor) Owner(ctx context.Context) (n graphqlbackend.NamespaceResolver, err error) {
	if m.namespaceOrgID == nil {
		n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, *m.namespaceUserID)
	} else {
		n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, *m.namespaceOrgID)
	}
	return n, err
}

func (m *monitor) Trigger(ctx context.Context) (graphqlbackend.MonitorTrigger, error) {
	q, err := m.triggerQueryByMonitorQuery(ctx, m.id)
	if err != nil {
		return nil, err
	}
	t, err := m.runTriggerQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	// Hydrate with resolver.
	t.(*monitorQuery).Resolver = m.Resolver
	return &monitorTrigger{t}, nil
}

func (m *monitor) Actions(ctx context.Context, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	return m.actionConnectionResolverWithTriggerID(ctx, nil, m.id, args)
}

func (r *Resolver) actionConnectionResolverWithTriggerID(ctx context.Context, triggerEventID *int64, monitorID int64, args *graphqlbackend.ListActionArgs) (c graphqlbackend.MonitorActionConnectionResolver, err error) {
	// For now, we only support emails as actions. Once we add other actions such as
	// webhooks, we have to query those tables here too.
	var q *sqlf.Query
	q, err = r.readActionEmailQuery(ctx, monitorID, args)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	es, err := scanEmails(rows)
	if err != nil {
		return nil, err
	}
	actions := make([]graphqlbackend.MonitorAction, 0, len(es))
	for _, e := range es {
		// Hydrate action with resolver.
		e.(*monitorEmail).Resolver = r
		actions = append(actions, &action{
			email: e,
		})
	}
	return &monitorActionConnection{
		actions:        actions,
		triggerEventID: triggerEventID,
	}, nil
}

//
// MonitorTrigger <<UNION>>
//
type monitorTrigger struct {
	query graphqlbackend.MonitorQueryResolver
}

func (t *monitorTrigger) ToMonitorQuery() (graphqlbackend.MonitorQueryResolver, bool) {
	return t.query, t.query != nil
}

//
// Query
//
type monitorQuery struct {
	*Resolver
	id        int64
	monitor   int64
	query     string
	createdBy int64
	createdAt time.Time
	changedBy int64
	changedAt time.Time
}

func (q *monitorQuery) ID() graphql.ID {
	return relay.MarshalID(monitorTriggerQueryKind, q.id)
}

func (q *monitorQuery) Query() string {
	return q.query
}

func (q *monitorQuery) Events(ctx context.Context, args *graphqlbackend.ListEventsArgs) graphqlbackend.MonitorTriggerEventConnectionResolver {
	return &monitorTriggerEventConnection{monitorID: relay.MarshalID(monitorKind, q.monitor), userID: relay.MarshalID("User", actor.FromContext(ctx).UID)}
}

//
// MonitorTriggerEventConnection
//
type monitorTriggerEventConnection struct {
	*Resolver
	monitorID graphql.ID
	userID    graphql.ID // TODO: remove this. Just for stub implementation
}

func (a *monitorTriggerEventConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorTriggerEventResolver, error) {
	return []graphqlbackend.MonitorTriggerEventResolver{&monitorTriggerEvent{
		Resolver:  a.Resolver,
		id:        42,
		status:    "SUCCESS",
		message:   nil,
		timestamp: graphqlbackend.DateTime{Time: time.Now()},
		monitorID: a.monitorID,
		userID:    a.userID,
	}}, nil
}

func (a *monitorTriggerEventConnection) TotalCount(ctx context.Context) (int32, error) {
	return 1, nil
}

func (a *monitorTriggerEventConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

//
// MonitorTriggerEvent
//
type monitorTriggerEvent struct {
	*Resolver
	id        int64
	status    string
	message   *string
	timestamp graphqlbackend.DateTime
	monitorID graphql.ID

	userID graphql.ID // TODO: remove this. Just for stub implementation
}

func (m *monitorTriggerEvent) ID() graphql.ID {
	return relay.MarshalID(monitorTriggerEventKind, m.id)
}

func (m *monitorTriggerEvent) Status() string {
	return m.status
}

func (m *monitorTriggerEvent) Message() *string {
	return m.message
}

func (m *monitorTriggerEvent) Timestamp() graphqlbackend.DateTime {
	return m.timestamp
}

func (m *monitorTriggerEvent) Actions(ctx context.Context, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	return m.actionConnectionResolverWithTriggerID(ctx, &m.id, m.id, args)
}

// ActionConnection
//
type monitorActionConnection struct {
	actions []graphqlbackend.MonitorAction

	// triggerEventID is used to link action events to a trigger event
	triggerEventID *int64
}

func (a *monitorActionConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorAction, error) {
	return a.actions, nil
}

func (a *monitorActionConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(a.actions)), nil
}

func (a *monitorActionConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if len(a.actions) == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	last := a.actions[len(a.actions)-1]
	if email, ok := last.ToMonitorEmail(); ok {
		return graphqlutil.NextPageCursor(string(email.ID())), nil
	}
	return nil, fmt.Errorf("we only support email actions for now")
}

//
// Action <<UNION>>
//
type action struct {
	email graphqlbackend.MonitorEmailResolver
}

func (a *action) ToMonitorEmail() (graphqlbackend.MonitorEmailResolver, bool) {
	return a.email, a.email != nil
}

//
// Email
//
type monitorEmail struct {
	*Resolver
	id        int64
	monitor   int64
	enabled   bool
	priority  string
	header    string
	createdBy int32
	createdAt time.Time
	changedBy int32
	changedAt time.Time

	// If triggerEventID == nil, all events of this action will be returned.
	// Otherwise, only those events of this action which are related to the specified
	// trigger event will be returned.
	triggerEventID *graphql.ID
}

func (m *monitorEmail) Recipients(ctx context.Context, args *graphqlbackend.ListRecipientsArgs) (c graphqlbackend.MonitorActionEmailRecipientsConnectionResolver, err error) {
	q, err := m.readRecipientQuery(ctx, m.id, args)
	if err != nil {
		return nil, err
	}
	rows, err := m.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ms, err := scanRecipients(rows)
	if err != nil {
		return nil, err
	}
	var ns []graphqlbackend.NamespaceResolver
	for _, r := range ms {
		n := graphqlbackend.NamespaceResolver{}
		if r.namespaceOrgID == nil {
			n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, *r.namespaceUserID)
		} else {
			n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, *r.namespaceOrgID)
		}
		ns = append(ns, n)
	}

	// Since recipients can either be a user or an org it would be very tedious to
	// use the user-id or org-id of the last entry as a cursor for the next page. It
	// is easier to just use the id of the recipients table.
	var nextPageCursor string
	if len(ms) > 0 {
		nextPageCursor = string(relay.MarshalID(monitorActionEmailRecipientKind, ms[len(ms)-1].id))
	}
	return &monitorActionEmailRecipientsConnection{ns, nextPageCursor}, nil
}

func (m *monitorEmail) Enabled() bool {
	return m.enabled
}

func (m *monitorEmail) Priority() string {
	return m.priority
}

func (m *monitorEmail) Header() string {
	return m.header
}

func (m *monitorEmail) ID() graphql.ID {
	return relay.MarshalID(monitorActionEmailKind, m.id)
}

func (m *monitorEmail) Events(ctx context.Context, args *graphqlbackend.ListEventsArgs) (graphqlbackend.MonitorActionEventConnectionResolver, error) {
	return &monitorActionEventConnection{}, nil
}

//
// MonitorActionEmailRecipientConnection
//
type monitorActionEmailRecipientsConnection struct {
	recipients     []graphqlbackend.NamespaceResolver
	nextPageCursor string
}

func (a *monitorActionEmailRecipientsConnection) Nodes(ctx context.Context) ([]graphqlbackend.NamespaceResolver, error) {
	return a.recipients, nil
}

func (a *monitorActionEmailRecipientsConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(a.recipients)), nil
}

func (a *monitorActionEmailRecipientsConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if len(a.recipients) == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(a.nextPageCursor), nil
}

//
// MonitorActionEventConnection
//
type monitorActionEventConnection struct {
}

func (a *monitorActionEventConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorActionEventResolver, error) {
	notImplemented := "message not implemented"
	return []graphqlbackend.MonitorActionEventResolver{
			&monitorActionEvent{id: "314", status: "SUCCESS", timestamp: graphqlbackend.DateTime{Time: time.Now()}},
			&monitorActionEvent{id: "315", status: "ERROR", message: &notImplemented, timestamp: graphqlbackend.DateTime{Time: time.Now()}},
		},
		nil
}

func (a *monitorActionEventConnection) TotalCount(ctx context.Context) (int32, error) {
	return 1, nil
}

func (a *monitorActionEventConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

//
// MonitorEvent
//
type monitorActionEvent struct {
	id        graphql.ID
	status    string
	message   *string
	timestamp graphqlbackend.DateTime
}

func (m *monitorActionEvent) ID() graphql.ID {
	return m.id
}

func (m *monitorActionEvent) Status() string {
	return m.status
}

func (m *monitorActionEvent) Message() *string {
	return m.message
}

func (m *monitorActionEvent) Timestamp() graphqlbackend.DateTime {
	return m.timestamp
}

func unmarshallAfter(after *string) (int64, error) {
	var a int64
	if after == nil {
		a = 0
	} else {
		err := relay.UnmarshalSpec(graphql.ID(*after), &a)
		if err != nil {
			return -1, err
		}
	}
	return a, nil
}

package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
)

// NewResolver returns a new Resolver who uses the given db
func NewResolver(db *sql.DB) graphqlbackend.CodeMonitorsResolver {
	return &Resolver{db: db, clock: func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }}
}

type Resolver struct {
	db    *sql.DB
	clock func() time.Time
}

func (r *Resolver) Monitors(ctx context.Context, userID int32, args *graphqlbackend.ListMonitorsArgs) (graphqlbackend.MonitorConnectionResolver, error) {
	q, err := monitorsQuery(userID, args)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ms, err := scanMonitors(rows)
	if err != nil {
		return nil, err
	}
	return &monitorConnection{r, ms}, nil
}

func (r *Resolver) CreateCodeMonitor(ctx context.Context, args *graphqlbackend.CreateCodeMonitorArgs) (graphqlbackend.MonitorResolver, error) {
	q, err := r.createCodeMonitorQuery(ctx, args)
	if err != nil {
		return nil, err
	}
	return r.runQuery(ctx, q)
}

func (r *Resolver) runQuery(ctx context.Context, q *sqlf.Query) (graphqlbackend.MonitorResolver, error) {
	rows, err := r.db.Query(q.Query(sqlf.PostgresBindVar), q.Args()...)
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
LIMIT %S
`
	var after int64
	if args.After == nil {
		after = 0
	} else {
		err := relay.UnmarshalSpec(graphql.ID(*args.After), &after)
		if err != nil {
			return nil, err
		}
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
(created_at, created_by, changed_at, changed_by, description, namespace_user_id, namespace_org_id) 
VALUES (%s,%s,%s,%s,%s,%s,%s)
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
		nilOrInt32(userID),
		nilOrInt32(orgID),
		sqlf.Join(monitorColumns, ", "),
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
	monitorKind = "codemonitor"
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
	return &monitorTrigger{&monitorQuery{monitorID: m.ID(), userID: m.ID()}}, nil
}

func (m *monitor) Actions(ctx context.Context, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	return &monitorActionConnection{
			monitorID: m.ID(),
			userID:    m.ID()},
		nil
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
	monitorID graphql.ID
	userID    graphql.ID // TODO: remove this. Just for stub implementation
}

func (q *monitorQuery) ID() graphql.ID {
	return "monitorQuery ID not implemented"
}

func (q *monitorQuery) Query() string {
	return "repo:github.com/sourcegraph/sourcegraph file:code_monitors not implemented"
}

func (q *monitorQuery) Events(ctx context.Context, args *graphqlbackend.ListEventsArgs) graphqlbackend.MonitorTriggerEventConnectionResolver {
	return &monitorTriggerEventConnection{monitorID: q.monitorID, userID: q.userID}
}

//
// MonitorTriggerEventConnection
//
type monitorTriggerEventConnection struct {
	monitorID graphql.ID
	userID    graphql.ID // TODO: remove this. Just for stub implementation
}

func (a *monitorTriggerEventConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorTriggerEventResolver, error) {
	return []graphqlbackend.MonitorTriggerEventResolver{&monitorTriggerEvent{
		id:        "42",
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
	id        graphql.ID
	status    string
	message   *string
	timestamp graphqlbackend.DateTime
	monitorID graphql.ID

	userID graphql.ID // TODO: remove this. Just for stub implementation
}

func (m *monitorTriggerEvent) ID() graphql.ID {
	return m.id
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
	return graphqlbackend.MonitorActionConnectionResolver(&monitorActionConnection{userID: m.userID, monitorID: m.monitorID, triggerEventID: &m.id}), nil
}

// ActionConnection
//
type monitorActionConnection struct {
	userID    graphql.ID //  TODO: remove this. This is just for the stub implementation.
	monitorID graphql.ID

	// triggerEventID is used to link action events to a trigger event
	triggerEventID *graphql.ID
}

func (a *monitorActionConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorAction, error) {
	return []graphqlbackend.MonitorAction{&action{email: &monitorEmail{id: "42", userID: a.userID}}}, nil
}

func (a *monitorActionConnection) TotalCount(ctx context.Context) (int32, error) {
	return 1, nil
}

func (a *monitorActionConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
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
	userID graphql.ID //  TODO: remove this. This is just for the stub implementation.
	id     graphql.ID

	// If triggerEventID == nil, all events of this action will be returned.
	// Otherwise, only those events of this action which are related to the specified
	// trigger event will be returned.
	triggerEventID *graphql.ID
}

func (m *monitorEmail) Recipient(ctx context.Context) (graphqlbackend.MonitorEmailRecipient, error) {
	user, err := graphqlbackend.UserByID(ctx, m.userID)
	return &monitorEmailRecipient{
		user: user,
	}, err
}

func (m *monitorEmail) Enabled() bool {
	return true
}

func (m *monitorEmail) Priority() string {
	return "NORMAL"
}

func (m *monitorEmail) Header() string {
	return "Header not implemented"
}

func (m *monitorEmail) ID() graphql.ID {
	return "monitorEmail ID not implemented"
}

func (m *monitorEmail) Events(ctx context.Context, args *graphqlbackend.ListEventsArgs) (graphqlbackend.MonitorActionEventConnectionResolver, error) {
	return &monitorActionEventConnection{}, nil
}

//
// MonitorEmailRecipient <<UNION>>
//
type MonitorEmailRecipient interface {
	ToUser() (*graphqlbackend.UserResolver, bool)
}

type monitorEmailRecipient struct {
	user *graphqlbackend.UserResolver
}

func (o *monitorEmailRecipient) ToUser() (*graphqlbackend.UserResolver, bool) {
	return o.user, o.user != nil
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

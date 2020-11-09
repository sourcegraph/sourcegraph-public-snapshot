package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type Resolver struct {
}

func (*Resolver) Monitors(ctx context.Context, userID graphql.ID, args *graphqlbackend.ListMonitorsArgs) (graphqlbackend.MonitorConnectionResolver, error) {
	return &monitorConnection{userID: userID}, nil
}

//
// MonitorConnection
//
type monitorConnection struct {
	userID graphql.ID
}

func (m *monitorConnection) Nodes(ctx context.Context) ([]graphqlbackend.MonitorResolver, error) {
	return []graphqlbackend.MonitorResolver{&monitor{userID: m.userID}}, nil
}

func (m *monitorConnection) TotalCount(ctx context.Context) (int32, error) {
	return 1, nil
}

func (m *monitorConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

//
// Monitor
//
type monitor struct {
	userID graphql.ID
}

func (m *monitor) ID() graphql.ID {
	return "ID not implemented"
}

func (m *monitor) CreatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByID(ctx, m.userID)
}

func (m *monitor) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: time.Now()}
}

func (m *monitor) Description() string {
	return "description not implemented"
}

func (*monitor) Enabled() bool {
	return true
}

func (m *monitor) Trigger(ctx context.Context) (graphqlbackend.MonitorTrigger, error) {
	return &monitorTrigger{&monitorQuery{monitorID: m.ID(), userID: m.userID}}, nil
}

func (m *monitor) Actions(ctx context.Context, args *graphqlbackend.ListActionArgs) (graphqlbackend.MonitorActionConnectionResolver, error) {
	return &monitorActionConnection{
			monitorID: m.ID(),
			userID:    m.userID},
		nil
}

func (m *monitor) Owner(ctx context.Context) (n graphqlbackend.NamespaceResolver, err error) {
	n.Namespace, err = graphqlbackend.UserByID(ctx, m.userID)
	return n, err
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

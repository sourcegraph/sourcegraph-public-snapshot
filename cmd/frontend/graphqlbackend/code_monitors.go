package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

//
// MonitorConnection
//
type ListMonitorsArgs struct {
	First int32
	After *string
}

func monitors(ctx context.Context, userID graphql.ID, args *ListMonitorsArgs) (MonitorConnectionResolver, error) {
	return &monitorConnection{userID: userID}, nil
}

type MonitorConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type monitorConnection struct {
	userID graphql.ID
}

func (m *monitorConnection) Nodes(ctx context.Context) ([]MonitorResolver, error) {
	return []MonitorResolver{&monitor{userID: m.userID}}, nil
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
type ListActionArgs struct {
	First int32
	After *string
}

type MonitorResolver interface {
	ID() graphql.ID
	CreatedBy(ctx context.Context) (*UserResolver, error)
	CreatedAt() DateTime
	Description() string
	Owner(ctx context.Context) (NamespaceResolver, error)
	Enabled() bool
	Trigger(ctx context.Context) (MonitorTrigger, error)
	Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error)
}

type monitor struct {
	userID graphql.ID
}

func (m *monitor) ID() graphql.ID {
	return graphql.ID("ID not implemented")
}

func (m *monitor) CreatedBy(ctx context.Context) (*UserResolver, error) {
	return UserByID(ctx, m.userID)
}

func (m *monitor) CreatedAt() DateTime {
	return DateTime{time.Now()}
}

func (m *monitor) Description() string {
	return "description not implemented"
}

func (*monitor) Enabled() bool {
	return true
}

func (m *monitor) Trigger(ctx context.Context) (MonitorTrigger, error) {
	return &monitorTrigger{&monitorQuery{monitorID: m.ID(), userID: m.userID}}, nil
}

func (m *monitor) Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error) {
	return &monitorActionConnection{
			monitorID: m.ID(),
			userID:    m.userID},
		nil
}

func (m *monitor) Owner(ctx context.Context) (n NamespaceResolver, err error) {
	n.Namespace, err = UserByID(ctx, m.userID)
	return n, err
}

//
// MonitorTrigger <<UNION>>
//
type MonitorTrigger interface {
	ToMonitorQuery() (MonitorQueryResolver, bool)
}

type monitorTrigger struct {
	query MonitorQueryResolver
}

func (t *monitorTrigger) ToMonitorQuery() (MonitorQueryResolver, bool) {
	return t.query, t.query != nil
}

//
// Query
//
type MonitorQueryResolver interface {
	ID() graphql.ID
	Query() string
	Events(ctx context.Context, args *ListEventsArgs) MonitorTriggerEventConnectionResolver
}

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

func (q *monitorQuery) Events(ctx context.Context, args *ListEventsArgs) MonitorTriggerEventConnectionResolver {
	return &monitorTriggerEventConnection{monitorID: q.monitorID, userID: q.userID}
}

//
// MonitorTriggerEventConnection
//
type MonitorTriggerEventConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorTriggerEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type monitorTriggerEventConnection struct {
	monitorID graphql.ID
	userID    graphql.ID // TODO: remove this. Just for stub implementation
}

func (a *monitorTriggerEventConnection) Nodes(ctx context.Context) ([]MonitorTriggerEventResolver, error) {
	return []MonitorTriggerEventResolver{&monitorTriggerEvent{
		id:        "42",
		status:    "SUCCESS",
		message:   nil,
		timestamp: DateTime{time.Now()},
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
type MonitorTriggerEventResolver interface {
	ID() graphql.ID
	Status() string
	Message() *string
	Timestamp() DateTime
	Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error)
}

type monitorTriggerEvent struct {
	id        graphql.ID
	status    string
	message   *string
	timestamp DateTime
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

func (m *monitorTriggerEvent) Timestamp() DateTime {
	return m.timestamp
}

func (m *monitorTriggerEvent) Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error) {
	return MonitorActionConnectionResolver(&monitorActionConnection{userID: m.userID, monitorID: m.monitorID, triggerEventID: &m.id}), nil
}

// ActionConnection
//
type MonitorActionConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorAction, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type monitorActionConnection struct {
	userID    graphql.ID //  TODO: remove this. This is just for the stub implementation.
	monitorID graphql.ID

	// triggerEventID is used to link action events to a trigger event
	triggerEventID *graphql.ID
}

func (a *monitorActionConnection) Nodes(ctx context.Context) ([]MonitorAction, error) {
	return []MonitorAction{&action{email: &monitorEmail{id: "42", userID: a.userID}}}, nil
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
type MonitorAction interface {
	ToMonitorEmail() (MonitorEmailResolver, bool)
}

type action struct {
	email MonitorEmailResolver
}

func (a *action) ToMonitorEmail() (MonitorEmailResolver, bool) {
	return a.email, a.email != nil
}

//
// Email
//
type MonitorEmailResolver interface {
	ID() graphql.ID
	Enabled() bool
	Priority() string
	Header() string
	Recipient(ctx context.Context) (MonitorEmailRecipient, error)
	Events(ctx context.Context, args *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type monitorEmail struct {
	userID graphql.ID //  TODO: remove this. This is just for the stub implementation.
	id     graphql.ID

	// If triggerEventID == nil, all events of this action will be returned.
	// Otherwise, only those events of this action which are related to the specified
	// trigger event will be returned.
	triggerEventID *graphql.ID
}

func (m *monitorEmail) Recipient(ctx context.Context) (MonitorEmailRecipient, error) {
	user, err := UserByID(ctx, m.userID)
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

func (m *monitorEmail) Events(ctx context.Context, args *ListEventsArgs) (MonitorActionEventConnectionResolver, error) {
	return &monitorActionEventConnection{}, nil
}

//
// MonitorEmailRecipient <<UNION>>
//
type MonitorEmailRecipient interface {
	ToUser() (*UserResolver, bool)
}

type monitorEmailRecipient struct {
	user *UserResolver
}

func (o *monitorEmailRecipient) ToUser() (*UserResolver, bool) {
	return o.user, o.user != nil
}

//
// MonitorActionEventConnection
//
type MonitorActionEventConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorActionEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type monitorActionEventConnection struct {
}

func (a *monitorActionEventConnection) Nodes(ctx context.Context) ([]MonitorActionEventResolver, error) {
	notImplemented := "message not implemented"
	return []MonitorActionEventResolver{
			&monitorActionEvent{id: "314", status: "SUCCESS", timestamp: DateTime{time.Now()}},
			&monitorActionEvent{id: "315", status: "ERROR", message: &notImplemented, timestamp: DateTime{time.Now()}},
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
type ListEventsArgs struct {
	First int32
	After *string
}

type MonitorActionEventResolver interface {
	ID() graphql.ID
	Status() string
	Message() *string
	Timestamp() DateTime
}

type monitorActionEvent struct {
	id        graphql.ID
	status    string
	message   *string
	timestamp DateTime
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

func (m *monitorActionEvent) Timestamp() DateTime {
	return m.timestamp
}

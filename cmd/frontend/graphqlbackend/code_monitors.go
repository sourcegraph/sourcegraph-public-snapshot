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
	Owner(ctx context.Context) (Owner, error)
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

func (m *monitor) Trigger(ctx context.Context) (MonitorTrigger, error) {
	return &monitorTrigger{&monitorQuery{monitorID: m.ID()}}, nil
}

func (m *monitor) Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error) {
	// TODO: fetch data
	return &monitorActionConnection{
			monitorID: m.ID(),
			userID:    m.userID}, // TODO: remove this. This is just for the stub implementation.
		nil
}

//
// Owner <<UNION>>
//
type Owner interface {
	ToUser() (*UserResolver, bool)
	ToOrg() (*OrgResolver, bool)
}

func (m *monitor) Owner(ctx context.Context) (Owner, error) {
	user, err := UserByID(ctx, m.userID)
	return &owner{user: user}, err
}

type owner struct {
	user *UserResolver
	org  *OrgResolver
}

func (o *owner) ToUser() (*UserResolver, bool) {
	return o.user, o.user != nil
}

func (o *owner) ToOrg() (*OrgResolver, bool) {
	return o.org, o.org != nil
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
}

type monitorQuery struct {
	monitorID graphql.ID
}

func (q *monitorQuery) ID() graphql.ID {
	return "monitorQuery ID not implemented"
}

func (q *monitorQuery) Query() string {
	return "repo:github.com/sourcegraph/sourcegraph file:code_monitors not implemented"
}

//
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
}

type monitorEmail struct {
	userID graphql.ID //  TODO: remove this. This is just for the stub implementation.
	id     graphql.ID
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

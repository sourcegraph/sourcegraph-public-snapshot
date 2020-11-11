package graphqlbackend

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type CodeMonitorsResolver interface {
	Monitors(ctx context.Context, userID int32, args *ListMonitorsArgs) (MonitorConnectionResolver, error)
	CreateCodeMonitor(ctx context.Context, args *CreateCodeMonitorArgs) (MonitorResolver, error)
}

type MonitorConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
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

type MonitorTrigger interface {
	ToMonitorQuery() (MonitorQueryResolver, bool)
}

type MonitorQueryResolver interface {
	ID() graphql.ID
	Query() string
	Events(ctx context.Context, args *ListEventsArgs) MonitorTriggerEventConnectionResolver
}

type MonitorTriggerEventConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorTriggerEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type MonitorTriggerEventResolver interface {
	ID() graphql.ID
	Status() string
	Message() *string
	Timestamp() DateTime
	Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error)
}

type MonitorActionConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorAction, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type MonitorAction interface {
	ToMonitorEmail() (MonitorEmailResolver, bool)
}

type MonitorEmailResolver interface {
	ID() graphql.ID
	Enabled() bool
	Priority() string
	Header() string
	Recipient(ctx context.Context) (MonitorEmailRecipient, error)
	Events(ctx context.Context, args *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorEmailRecipient interface {
	ToUser() (*UserResolver, bool)
}

type MonitorActionEventConnectionResolver interface {
	Nodes(ctx context.Context) ([]MonitorActionEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type MonitorActionEventResolver interface {
	ID() graphql.ID
	Status() string
	Message() *string
	Timestamp() DateTime
}

type ListEventsArgs struct {
	First int32
	After *string
}

type ListMonitorsArgs struct {
	First int32
	After *string
}

type ListActionArgs struct {
	First int32
	After *string
}

type CreateCodeMonitorArgs struct {
	Namespace   graphql.ID
	Description string
	Enabled     bool
	Trigger     *CreateTriggerArgs
	Actions     []*CreateActionArgs
}

type CreateTriggerArgs struct {
	Query string
}

type CreateActionArgs struct {
	Email *CreateActionEmailArgs
}

type CreateActionEmailArgs struct {
	Enabled    bool
	Priority   string
	Recipients []graphql.ID
	Header     string
}

var DefaultCodeMonitorsResolver = &defaultCodeMonitorsResolver{}

var codeMonitorsOnlyInEnterprise = errors.New("code monitors are only available in enterprise")

type defaultCodeMonitorsResolver struct {
}

func (d defaultCodeMonitorsResolver) Monitors(ctx context.Context, userID int32, args *ListMonitorsArgs) (MonitorConnectionResolver, error) {
	return nil, codeMonitorsOnlyInEnterprise
}

func (d defaultCodeMonitorsResolver) CreateCodeMonitor(ctx context.Context, args *CreateCodeMonitorArgs) (MonitorResolver, error) {
	return nil, codeMonitorsOnlyInEnterprise
}

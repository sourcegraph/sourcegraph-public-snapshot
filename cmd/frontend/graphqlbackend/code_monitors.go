package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func (s *schemaResolver) Monitors(ctx context.Context, args *ListMonitorsArgs) (MonitorConnectionResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, s.db); err != nil {
		return nil, err
	}

	return s.CodeMonitorsResolver.Monitors(ctx, nil, args)
}

type CodeMonitorsResolver interface {
	// Query
	Monitors(ctx context.Context, userID *int32, args *ListMonitorsArgs) (MonitorConnectionResolver, error)
	MonitorByID(ctx context.Context, id graphql.ID) (MonitorResolver, error)

	// Mutations
	CreateCodeMonitor(ctx context.Context, args *CreateCodeMonitorArgs) (MonitorResolver, error)
	ToggleCodeMonitor(ctx context.Context, args *ToggleCodeMonitorArgs) (MonitorResolver, error)
	DeleteCodeMonitor(ctx context.Context, args *DeleteCodeMonitorArgs) (*EmptyResponse, error)
	UpdateCodeMonitor(ctx context.Context, args *UpdateCodeMonitorArgs) (MonitorResolver, error)
	ResetTriggerQueryTimestamps(ctx context.Context, args *ResetTriggerQueryTimestampsArgs) (*EmptyResponse, error)
	TriggerTestEmailAction(ctx context.Context, args *TriggerTestEmailActionArgs) (*EmptyResponse, error)
	TriggerTestWebhookAction(ctx context.Context, args *TriggerTestWebhookActionArgs) (*EmptyResponse, error)
	TriggerTestSlackWebhookAction(ctx context.Context, args *TriggerTestSlackWebhookActionArgs) (*EmptyResponse, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type MonitorConnectionResolver interface {
	Nodes() []MonitorResolver
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type MonitorResolver interface {
	ID() graphql.ID
	CreatedBy(ctx context.Context) (*UserResolver, error)
	CreatedAt() gqlutil.DateTime
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
	Events(ctx context.Context, args *ListEventsArgs) (MonitorTriggerEventConnectionResolver, error)
}

type MonitorTriggerEventConnectionResolver interface {
	Nodes() []MonitorTriggerEventResolver
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type MonitorTriggerEventResolver interface {
	ID() graphql.ID
	Status() (string, error)
	Message() *string
	Timestamp() (gqlutil.DateTime, error)
	Actions(ctx context.Context, args *ListActionArgs) (MonitorActionConnectionResolver, error)
	ResultCount() int32
	Query() *string
}

type MonitorActionConnectionResolver interface {
	Nodes() []MonitorAction
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type MonitorAction interface {
	ID() graphql.ID
	ToMonitorEmail() (MonitorEmailResolver, bool)
	ToMonitorWebhook() (MonitorWebhookResolver, bool)
	ToMonitorSlackWebhook() (MonitorSlackWebhookResolver, bool)
}

type MonitorEmailResolver interface {
	ID() graphql.ID
	Enabled() bool
	IncludeResults() bool
	Priority() string
	Header() string
	Recipients(ctx context.Context, args *ListRecipientsArgs) (MonitorActionEmailRecipientsConnectionResolver, error)
	Events(ctx context.Context, args *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorWebhookResolver interface {
	ID() graphql.ID
	Enabled() bool
	IncludeResults() bool
	URL() string
	Events(ctx context.Context, args *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorSlackWebhookResolver interface {
	ID() graphql.ID
	Enabled() bool
	IncludeResults() bool
	URL() string
	Events(ctx context.Context, args *ListEventsArgs) (MonitorActionEventConnectionResolver, error)
}

type MonitorEmailRecipient interface {
	ToUser() (*UserResolver, bool)
}

type MonitorActionEmailRecipientsConnectionResolver interface {
	Nodes() []NamespaceResolver
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type MonitorActionEventConnectionResolver interface {
	Nodes() []MonitorActionEventResolver
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type MonitorActionEventResolver interface {
	ID() graphql.ID
	Status() (string, error)
	Message() *string
	Timestamp() gqlutil.DateTime
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

type ListRecipientsArgs struct {
	First int32
	After *string
}

type CreateCodeMonitorArgs struct {
	Monitor *CreateMonitorArgs
	Trigger *CreateTriggerArgs
	Actions []*CreateActionArgs
}

type CreateTriggerArgs struct {
	Query string
}

type CreateActionArgs struct {
	Email        *CreateActionEmailArgs
	Webhook      *CreateActionWebhookArgs
	SlackWebhook *CreateActionSlackWebhookArgs
}

type CreateActionEmailArgs struct {
	Enabled        bool
	IncludeResults bool
	Priority       string
	Recipients     []graphql.ID
	Header         string
}

type CreateActionWebhookArgs struct {
	Enabled        bool
	IncludeResults bool
	URL            string
}

type CreateActionSlackWebhookArgs struct {
	Enabled        bool
	IncludeResults bool
	URL            string
}

type ToggleCodeMonitorArgs struct {
	Id      graphql.ID
	Enabled bool
}

type DeleteCodeMonitorArgs struct {
	Id graphql.ID
}

type ResetTriggerQueryTimestampsArgs struct {
	Id graphql.ID
}

type TriggerTestEmailActionArgs struct {
	Namespace   graphql.ID
	Description string
	Email       *CreateActionEmailArgs
}

type TriggerTestWebhookActionArgs struct {
	Namespace   graphql.ID
	Description string
	Webhook     *CreateActionWebhookArgs
}

type TriggerTestSlackWebhookActionArgs struct {
	Namespace    graphql.ID
	Description  string
	SlackWebhook *CreateActionSlackWebhookArgs
}

type CreateMonitorArgs struct {
	Namespace   graphql.ID
	Description string
	Enabled     bool
}

type EditActionEmailArgs struct {
	Id     *graphql.ID
	Update *CreateActionEmailArgs
}

type EditActionWebhookArgs struct {
	Id     *graphql.ID
	Update *CreateActionWebhookArgs
}

type EditActionSlackWebhookArgs struct {
	Id     *graphql.ID
	Update *CreateActionSlackWebhookArgs
}

type EditActionArgs struct {
	Email        *EditActionEmailArgs
	Webhook      *EditActionWebhookArgs
	SlackWebhook *EditActionSlackWebhookArgs
}

type EditTriggerArgs struct {
	Id     graphql.ID
	Update *CreateTriggerArgs
}

type EditMonitorArgs struct {
	Id     graphql.ID
	Update *CreateMonitorArgs
}

type UpdateCodeMonitorArgs struct {
	Monitor *EditMonitorArgs
	Trigger *EditTriggerArgs
	Actions []*EditActionArgs
}

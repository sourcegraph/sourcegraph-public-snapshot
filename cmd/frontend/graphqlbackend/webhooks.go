package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// WebhooksResolver is a main interface for all GraphQL operations with webhooks.
type WebhooksResolver interface {
	CreateWebhook(ctx context.Context, args *CreateWebhookArgs) (WebhookResolver, error)
	DeleteWebhook(ctx context.Context, args *DeleteWebhookArgs) (*EmptyResponse, error)
	UpdateWebhook(ctx context.Context, args *UpdateWebhookArgs) (WebhookResolver, error)
	Webhooks(ctx context.Context, args *ListWebhookArgs) (WebhookConnectionResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

// WebhookConnectionResolver is an interface for querying lists of webhooks.
type WebhookConnectionResolver interface {
	Nodes(ctx context.Context) ([]WebhookResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*gqlutil.PageInfo, error)
}

// WebhookResolver is an interface for querying a single webhook.
type WebhookResolver interface {
	ID() graphql.ID
	UUID() string
	URL() (string, error)
	Name() string
	CodeHostURN() string
	CodeHostKind() string
	Secret(ctx context.Context) (*string, error)
	CreatedAt() gqlutil.DateTime
	UpdatedAt() gqlutil.DateTime
	CreatedBy(ctx context.Context) (*UserResolver, error)
	UpdatedBy(ctx context.Context) (*UserResolver, error)
	WebhookLogs(ctx context.Context, args *WebhookLogsArgs) (*WebhookLogConnectionResolver, error)
}

type CreateWebhookArgs struct {
	Name         string
	CodeHostKind string
	CodeHostURN  string
	Secret       *string
}

type DeleteWebhookArgs struct {
	ID graphql.ID
}

type UpdateWebhookArgs struct {
	ID           graphql.ID
	Name         *string
	CodeHostKind *string
	CodeHostURN  *string
	Secret       *string
}

type ListWebhookArgs struct {
	gqlutil.ConnectionArgs
	After *string
	Kind  *string
}

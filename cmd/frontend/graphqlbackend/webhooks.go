package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// WebhooksResolver is a main interface for all GraphQL operations with webhooks.
type WebhooksResolver interface {
	CreateWebhook(ctx context.Context, args *CreateWebhookArgs) (WebhookResolver, error)
	DeleteWebhook(ctx context.Context, args *DeleteWebhookArgs) (*EmptyResponse, error)
	Webhooks(ctx context.Context, args *ListWebhookArgs) (WebhookConnectionResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

// WebhookConnectionResolver is an interface for querying lists of webhooks.
type WebhookConnectionResolver interface {
	Nodes(ctx context.Context) ([]WebhookResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

// WebhookResolver is an interface for querying a single webhook.
type WebhookResolver interface {
	ID() graphql.ID
	UUID() string
	URL() (string, error)
	CodeHostURN() string
	CodeHostKind() string
	Secret(ctx context.Context) (*string, error)
	CreatedAt() gqlutil.DateTime
	UpdatedAt() gqlutil.DateTime
	CreatedBy(ctx context.Context) (*UserResolver, error)
	UpdatedBy(ctx context.Context) (*UserResolver, error)
}

type CreateWebhookArgs struct {
	CodeHostKind string
	CodeHostURN  string
	Secret       *string
}

type DeleteWebhookArgs struct {
	ID graphql.ID
}

type ListWebhookArgs struct {
	graphqlutil.ConnectionArgs
	After *string
	Kind  *string
}

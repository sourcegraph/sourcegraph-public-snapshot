package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO: Tests

type webhookResolver struct {
	db   database.DB
	hook *types.Webhook
}

func (r *webhookResolver) ID() graphql.ID {
	return marshalWebhookID(r.hook.ID)
}

func (r *webhookResolver) UUID() string {
	return r.hook.UUID.String()
}

func (r *webhookResolver) URL() (string, error) {
	externalURL, err := url.Parse(conf.Get().ExternalURL)
	if err != nil {
		return "", errors.Wrap(err, "could not parse site config external URL")
	}
	externalURL.Path = fmt.Sprintf("webhooks/%v", r.hook.UUID)
	return externalURL.String(), nil
}

func (r *webhookResolver) CodeHostURN() string {
	return r.hook.CodeHostURN
}

func (r *webhookResolver) CodeHostKind() string {
	return r.hook.CodeHostKind
}

func (r *webhookResolver) Secret(ctx context.Context) (*string, error) {
	// Secret is optional
	if r.hook.Secret == nil {
		return nil, nil
	}
	s, err := r.hook.Secret.Decrypt(ctx)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *webhookResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.hook.CreatedAt}
}

func (r *webhookResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.hook.UpdatedAt}
}

func (r *schemaResolver) Webhooks(ctx context.Context, args *struct {
	First *int    // Default to 20
	After *string // Default to first item
	Kind  *string // Default to no filtering
}) *webhookConnectionResolver {
	// TODO: Use the fields above to fetch the list of desired hooks
	return &webhookConnectionResolver{}
}

type webhookConnectionResolver struct {
}

func (r *webhookConnectionResolver) Nodes() ([]*webhookResolver, error) {
	return nil, errors.New("TODO: Nodes")
}

func (r *webhookConnectionResolver) TotalCount() (int32, error) {
	return 0, errors.New("TODO: TotalCount")
}

func (r *webhookConnectionResolver) PageInfo() (*graphqlutil.PageInfo, error) {
	return nil, errors.New("TODO: PageInfo")
}

func webhookByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*webhookResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalWebhookID(gqlID)
	if err != nil {
		return nil, err
	}

	hook, err := db.Webhooks(keyring.Default().WebhookLogKey).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &webhookResolver{db: db, hook: hook}, nil
}

func marshalWebhookID(id int32) graphql.ID {
	return relay.MarshalID("Webhook", id)
}

func unmarshalWebhookID(id graphql.ID) (hookID int32, err error) {
	err = relay.UnmarshalSpec(id, &hookID)
	return
}

func (r *schemaResolver) CreateWebhook(ctx context.Context, args *struct {
	CodeHostKind string
	CodeHostURN  string
	Secret       *string
}) (*webhookResolver, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}
	ws := backend.NewWebhookService(r.db, keyring.Default())
	webhook, err := ws.CreateWebhook(ctx, args.CodeHostKind, args.CodeHostURN, args.Secret)
	if err != nil {
		return nil, err
	}
	return &webhookResolver{hook: webhook, db: r.db}, nil
}

func (r *schemaResolver) DeleteWebhook(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}
	id, err := unmarshalWebhookID(args.ID)
	if err != nil {
		return nil, err
	}
	err = r.db.Webhooks(keyring.Default().WebhookKey).Delete(ctx, database.DeleteWebhookOpts{ID: id})
	if err != nil {
		return nil, errors.Wrap(err, "delete webhook")
	}
	return &EmptyResponse{}, nil
}

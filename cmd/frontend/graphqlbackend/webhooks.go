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

func (r *schemaResolver) Webhooks(ctx context.Context, args *webhookArgs) (*webhookConnectionResolver, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}
	opts, err := args.toWebhookListOptions()
	if err != nil {
		return nil, err
	}
	return &webhookConnectionResolver{
		db:  r.db,
		opt: opts,
	}, nil
}

type webhookArgs struct {
	graphqlutil.ConnectionArgs
	After *string
	Kind  *string
}

func (args *webhookArgs) toWebhookListOptions() (database.WebhookListOptions, error) {
	opt := database.WebhookListOptions{}
	if args.Kind != nil {
		opt.Kind = *args.Kind
	}
	if args.After != nil {
		cursor, err := UnmarshalWebhookCursor(args.After)
		if err != nil {
			return opt, err
		}
		opt.Cursor = cursor
	} else {
		opt.Cursor = &types.Cursor{
			Column:    "id",
			Direction: "next",
		}
	}
	args.Set(&opt.LimitOffset)
	return opt, nil
}

type webhookConnectionResolver struct {
	db  database.DB
	opt database.WebhookListOptions
}

func (r *webhookConnectionResolver) Nodes(ctx context.Context) ([]*webhookResolver, error) {
	webhooks, err := r.db.Webhooks(keyring.Default().WebhookKey).List(ctx, r.opt)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*webhookResolver, 0, len(webhooks))
	for _, wh := range webhooks {
		resolvers = append(resolvers, &webhookResolver{
			hook: wh,
		})
	}
	return resolvers, nil
}

func (r *webhookConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// TODO: implement Count db method?
	webhooks, err := r.db.Webhooks(keyring.Default().WebhookKey).List(ctx, r.opt)
	if err != nil {
		return 0, err
	}
	return int32(len(webhooks)), nil
}

func (r *webhookConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	webhooks, err := r.db.Webhooks(keyring.Default().WebhookKey).List(ctx, r.opt)
	if err != nil {
		return nil, err
	}
	if len(webhooks) == 0 || r.opt.LimitOffset == nil || len(webhooks) <= r.opt.Limit || r.opt.Cursor == nil {
		return graphqlutil.HasNextPage(false), nil
	}

	value := webhooks[len(webhooks)-1].ID
	return graphqlutil.NextPageCursor(MarshalWebhookCursor(
		&types.Cursor{
			Column:    r.opt.Cursor.Column,
			Value:     string(value),
			Direction: r.opt.Cursor.Direction,
		},
	)), nil
}

func webhookByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*webhookResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalWebhookID(gqlID)
	if err != nil {
		return nil, err
	}

	hook, err := db.Webhooks(keyring.Default().WebhookKey).GetByID(ctx, id)
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

const webhookCursorKind = "WebhookCursor"

func MarshalWebhookCursor(cursor *types.Cursor) string {
	return string(relay.MarshalID(webhookCursorKind, cursor))
}

func UnmarshalWebhookCursor(cursor *string) (*types.Cursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relay.UnmarshalKind(graphql.ID(*cursor)); kind != webhookCursorKind {
		return nil, errors.Errorf("cannot unmarshal repository cursor type: %q", kind)
	}
	var spec *types.Cursor
	if err := relay.UnmarshalSpec(graphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

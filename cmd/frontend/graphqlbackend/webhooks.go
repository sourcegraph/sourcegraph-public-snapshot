package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"sync"

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
	return r.hook.CodeHostURN.String()
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

type WebhookConnectionResolver interface {
	Nodes(ctx context.Context) ([]*webhookResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

var _ WebhookConnectionResolver = &webhookConnectionResolver{}

type webhookConnectionResolver struct {
	db       database.DB
	opt      database.WebhookListOptions
	once     sync.Once
	webhooks []*types.Webhook
	next     int32
	err      error
}

func (r *webhookConnectionResolver) Nodes(ctx context.Context) ([]*webhookResolver, error) {
	webhooks, _, err := r.compute(ctx)
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
	count, err := r.db.Webhooks(keyring.Default().WebhookLogKey).Count(ctx, r.opt)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (r *webhookConnectionResolver) compute(ctx context.Context) ([]*types.Webhook, int32, error) {
	r.once.Do(func() {
		opts := copyOpts(r.opt)
		if r.opt.LimitOffset != nil {
			opts.Limit++
		}
		r.webhooks, r.err = r.db.Webhooks(keyring.Default().WebhookKey).List(ctx, opts)
		if r.opt.LimitOffset != nil && opts.Limit != 0 && len(r.webhooks) == opts.Limit {
			r.next = r.webhooks[len(r.webhooks)-1].ID
			r.webhooks = r.webhooks[:len(r.webhooks)-1]
		}
	})
	return r.webhooks, r.next, r.err
}

func copyOpts(opts database.WebhookListOptions) database.WebhookListOptions {
	copied := database.WebhookListOptions{
		Kind:   opts.Kind,
		Cursor: opts.Cursor,
	}
	if opts.LimitOffset != nil {
		limitOffset := database.LimitOffset{
			Limit:  opts.Limit,
			Offset: opts.Offset,
		}
		copied.LimitOffset = &limitOffset
	}
	return copied
}

func (r *webhookConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next == 0 {
		return graphqlutil.HasNextPage(false), nil
	}

	return graphqlutil.NextPageCursor(MarshalWebhookCursor(
		&types.Cursor{
			Column:    r.opt.Cursor.Column,
			Value:     fmt.Sprintf("%d", next),
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
		return nil, errors.Errorf("cannot unmarshal webhook cursor type: %q", kind)
	}
	var spec *types.Cursor
	if err := relay.UnmarshalSpec(graphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
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

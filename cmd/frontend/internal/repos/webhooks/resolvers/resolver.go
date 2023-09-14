package resolvers

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	webhookKind       = "Webhook"
	webhookCursorKind = "WebhookCursor"
)

var _ graphqlbackend.WebhooksResolver = &webhooksResolver{}

type webhooksResolver struct {
	db database.DB
}

func NewWebhooksResolver(db database.DB) graphqlbackend.WebhooksResolver {
	return &webhooksResolver{db: db}
}

func (r *webhooksResolver) CreateWebhook(ctx context.Context, args *graphqlbackend.CreateWebhookArgs) (graphqlbackend.WebhookResolver, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}
	ws := backend.NewWebhookService(r.db, keyring.Default())
	webhook, err := ws.CreateWebhook(ctx, args.Name, args.CodeHostKind, args.CodeHostURN, args.Secret)
	if err != nil {
		return nil, err
	}
	return &webhookResolver{hook: webhook, db: r.db}, nil
}

func (r *webhooksResolver) DeleteWebhook(ctx context.Context, args *graphqlbackend.DeleteWebhookArgs) (*graphqlbackend.EmptyResponse, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}

	id, err := UnmarshalWebhookID(args.ID)
	if err != nil {
		return nil, err
	}
	ws := backend.NewWebhookService(r.db, keyring.Default())
	err = ws.DeleteWebhook(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "delete webhook")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *webhooksResolver) UpdateWebhook(ctx context.Context, args *graphqlbackend.UpdateWebhookArgs) (graphqlbackend.WebhookResolver, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}

	whID, err := UnmarshalWebhookID(args.ID)
	if err != nil {
		return nil, err
	}

	ws := backend.NewWebhookService(r.db, keyring.Default())
	var name string
	if args.Name != nil {
		name = *args.Name
	}
	var codeHostKind string
	if args.CodeHostKind != nil {
		codeHostKind = *args.CodeHostKind
	}
	var codeHostURN string
	if args.CodeHostURN != nil {
		codeHostURN = *args.CodeHostURN
	}

	webhook, err := ws.UpdateWebhook(ctx, whID, name, codeHostKind, codeHostURN, args.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "update webhook")
	}

	return &webhookResolver{hook: webhook, db: r.db}, nil
}

func (r *webhooksResolver) Webhooks(ctx context.Context, args *graphqlbackend.ListWebhookArgs) (graphqlbackend.WebhookConnectionResolver, error) {
	if auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, auth.ErrMustBeSiteAdmin
	}
	opts, err := toWebhookListOptions(args)
	if err != nil {
		return nil, err
	}
	return &webhooksConnectionResolver{
		db:  r.db,
		opt: opts,
	}, nil
}

func (r *webhooksResolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		webhookKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return webhookByID(ctx, r.db, id)
		},
	}
}

func webhookByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*webhookResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := UnmarshalWebhookID(gqlID)
	if err != nil {
		return nil, err
	}

	hook, err := db.Webhooks(keyring.Default().WebhookKey).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &webhookResolver{db: db, hook: hook}, nil
}

func toWebhookListOptions(args *graphqlbackend.ListWebhookArgs) (database.WebhookListOptions, error) {
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

var _ graphqlbackend.WebhookConnectionResolver = &webhooksConnectionResolver{}

type webhooksConnectionResolver struct {
	db       database.DB
	opt      database.WebhookListOptions
	once     sync.Once
	webhooks []*types.Webhook
	next     int32
	err      error
}

func (c *webhooksConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.WebhookResolver, error) {
	webhooks, _, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.WebhookResolver, 0, len(webhooks))
	for _, wh := range webhooks {
		resolvers = append(resolvers, &webhookResolver{
			db:   c.db,
			hook: wh,
		})
	}
	return resolvers, nil
}

func (c *webhooksConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := c.db.Webhooks(keyring.Default().WebhookLogKey).Count(ctx, c.opt)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (c *webhooksConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next == 0 {
		return graphqlutil.HasNextPage(false), nil
	}

	return graphqlutil.NextPageCursor(MarshalWebhookCursor(
		&types.Cursor{
			Column:    c.opt.Cursor.Column,
			Value:     fmt.Sprintf("%d", next),
			Direction: c.opt.Cursor.Direction,
		},
	)), nil
}

func (c *webhooksConnectionResolver) compute(ctx context.Context) ([]*types.Webhook, int32, error) {
	c.once.Do(func() {
		opts := copyOpts(c.opt)
		if c.opt.LimitOffset != nil {
			opts.Limit++
		}
		c.webhooks, c.err = c.db.Webhooks(keyring.Default().WebhookKey).List(ctx, opts)
		if c.opt.LimitOffset != nil && opts.Limit != 0 && len(c.webhooks) == opts.Limit {
			c.next = c.webhooks[len(c.webhooks)-1].ID
			c.webhooks = c.webhooks[:len(c.webhooks)-1]
		}
	})
	return c.webhooks, c.next, c.err
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

var _ graphqlbackend.WebhookResolver = &webhookResolver{}

type webhookResolver struct {
	db   database.DB
	hook *types.Webhook
}

func NewWebhookResolver(db database.DB, hook *types.Webhook) *webhookResolver {
	return &webhookResolver{
		db:   db,
		hook: hook,
	}
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
	externalURL.Path = fmt.Sprintf(".api/webhooks/%v", r.hook.UUID)
	return externalURL.String(), nil
}

func (r *webhookResolver) Name() string {
	return r.hook.Name
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

func (r *webhookResolver) CreatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if r.hook.CreatedByUserID == 0 {
		return nil, nil
	}

	user, err := graphqlbackend.UserByIDInt32(ctx, r.db, r.hook.CreatedByUserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return user, err
}

func (r *webhookResolver) UpdatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if r.hook.UpdatedByUserID == 0 {
		return nil, nil
	}

	user, err := graphqlbackend.UserByIDInt32(ctx, r.db, r.hook.UpdatedByUserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return user, err
}

func (r *webhookResolver) WebhookLogs(ctx context.Context, args *graphqlbackend.WebhookLogsArgs) (*graphqlbackend.WebhookLogConnectionResolver, error) {
	gqlID := marshalWebhookID(r.hook.ID)
	// We need to make a new args struct, otherwise the pointer gets shared
	// between resolvers.
	resolverArgs := *args
	resolverArgs.WebhookID = &gqlID
	return graphqlbackend.NewWebhookLogConnectionResolver(ctx, r.db, &resolverArgs, graphqlbackend.WebhookLogsAllExternalServices)
}

func marshalWebhookID(id int32) graphql.ID {
	return relay.MarshalID("Webhook", id)
}

func UnmarshalWebhookID(id graphql.ID) (hookID int32, err error) {
	err = relay.UnmarshalSpec(id, &hookID)
	return
}

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

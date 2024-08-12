package graphqlbackend

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/webhooks/outbound"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const outboundWebhookIDKind = "OutboundWebhook"

type OutboundWebhookConnectionResolver interface {
	Nodes() ([]OutboundWebhookResolver, error)
	TotalCount() (int32, error)
	PageInfo() (*gqlutil.PageInfo, error)
}

type OutboundWebhookEventTypeResolver interface {
	Key() string
	Description() string
}

type OutboundWebhookResolver interface {
	ID() graphql.ID
	URL(context.Context) (string, error)
	EventTypes() ([]OutboundWebhookScopedEventTypeResolver, error)
	Stats(context.Context) (OutboundWebhookLogStatsResolver, error)
	Logs(context.Context, OutboundWebhookLogsArgs) (OutboundWebhookLogConnectionResolver, error)
}

type OutboundWebhookScopedEventTypeResolver interface {
	EventType() string
	Scope() *string
}

type CreateOutboundWebhookArgs struct {
	Input OutboundWebhookCreateInput `json:"input"`
}

type DeleteOutboundWebhookArgs struct {
	ID graphql.ID `json:"id"`
}

type UpdateOutboundWebhookArgs struct {
	ID    graphql.ID                 `json:"id"`
	Input OutboundWebhookUpdateInput `json:"input"`
}

type ListOutboundWebhooksArgs struct {
	First     int32   `json:"first"`
	After     *string `json:"after"`
	EventType *string `json:"eventType"`
	Scope     *string `json:"scope"`
}

type OutboundWebhookLogsArgs struct {
	First      int32   `json:"first"`
	After      *string `json:"after"`
	OnlyErrors *bool   `json:"onlyErrors"`
}

type OutboundWebhookCreateInput struct {
	OutboundWebhookUpdateInput
	Secret string `json:"secret"`
}

type OutboundWebhookScopedEventTypeInput struct {
	EventType string  `json:"eventType"`
	Scope     *string `json:"scope"`
}

type OutboundWebhookUpdateInput struct {
	URL        string                                `json:"url"`
	EventTypes []OutboundWebhookScopedEventTypeInput `json:"eventTypes"`
}

func (r *schemaResolver) OutboundWebhooks(ctx context.Context, args ListOutboundWebhooksArgs) (OutboundWebhookConnectionResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opts := database.OutboundWebhookListOpts{
		LimitOffset: &database.LimitOffset{
			Limit: int(args.First),
		},
	}
	if args.After != nil {
		offset, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, errors.Newf("cannot parse offset %q", *args.After)
		}
		opts.Offset = offset
	}
	if args.EventType != nil {
		opts.EventTypes = []database.FilterEventType{{EventType: *args.EventType, Scope: args.Scope}}
	}

	return newOutboundWebhookConnectionResolver(ctx, outboundWebhookStore(r.db), opts), nil
}

func (r *schemaResolver) OutboundWebhookEventTypes(ctx context.Context) ([]OutboundWebhookEventTypeResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	eventTypes := outbound.GetRegisteredEventTypes()
	sort.Slice(eventTypes, func(i, j int) bool {
		return eventTypes[i].Key < eventTypes[j].Key
	})
	resolvers := make([]OutboundWebhookEventTypeResolver, len(eventTypes))
	for i, et := range eventTypes {
		resolvers[i] = &outboundWebhookEventTypeResolver{et}
	}

	return resolvers, nil
}

func (r *schemaResolver) CreateOutboundWebhook(ctx context.Context, args CreateOutboundWebhookArgs) (OutboundWebhookResolver, error) {
	user, err := auth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if !user.SiteAdmin {
		return nil, auth.ErrMustBeSiteAdmin
	}

	// Validate the URL
	err = outbound.CheckURL(args.Input.URL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid webhook address")
	}

	webhook := &types.OutboundWebhook{
		CreatedBy:  user.ID,
		UpdatedBy:  user.ID,
		URL:        encryption.NewUnencrypted(args.Input.URL),
		Secret:     encryption.NewUnencrypted(args.Input.Secret),
		EventTypes: outboundWebhookEventTypes(args.Input.EventTypes),
	}

	store := outboundWebhookStore(r.db)
	if err := store.Create(ctx, webhook); err != nil {
		return nil, err
	}

	return newOutboundWebhookResolverFromWebhook(store, webhook), nil
}

func (r *schemaResolver) DeleteOutboundWebhook(ctx context.Context, args DeleteOutboundWebhookArgs) (*EmptyResponse, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmarshalOutboundWebhookID(args.ID)
	if err != nil {
		return nil, err
	}

	store := outboundWebhookStore(r.db)
	if err := store.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) UpdateOutboundWebhook(ctx context.Context, args UpdateOutboundWebhookArgs) (OutboundWebhookResolver, error) {
	user, err := auth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if !user.SiteAdmin {
		return nil, auth.ErrMustBeSiteAdmin
	}

	id, err := unmarshalOutboundWebhookID(args.ID)
	if err != nil {
		return nil, err
	}

	store, err := outboundWebhookStore(r.db).Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = store.Done(err) }()

	webhook, err := store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	webhook.UpdatedBy = user.ID
	webhook.URL = encryption.NewUnencrypted(args.Input.URL)
	webhook.EventTypes = outboundWebhookEventTypes(args.Input.EventTypes)

	if err := store.Update(ctx, webhook); err != nil {
		return nil, err
	}

	return newOutboundWebhookResolverFromWebhook(store, webhook), nil
}

func OutboundWebhookByID(ctx context.Context, db database.DB, gql graphql.ID) (OutboundWebhookResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalOutboundWebhookID(gql)
	if err != nil {
		return nil, err
	}

	return newOutboundWebhookResolverFromDatabase(ctx, outboundWebhookStore(db), id), nil
}

func marshalOutboundWebhookID(id int64) graphql.ID {
	return relay.MarshalID(outboundWebhookIDKind, id)
}

func unmarshalOutboundWebhookID(gql graphql.ID) (id int64, err error) {
	if kind := relay.UnmarshalKind(gql); kind != outboundWebhookIDKind {
		return 0, errors.Newf("invalid outbound webhook id of kind %q", kind)
	}

	err = relay.UnmarshalSpec(gql, &id)
	return
}

type outboundWebhookConnectionResolver struct {
	nodes      func() ([]*types.OutboundWebhook, error)
	resolvers  func() ([]OutboundWebhookResolver, error)
	totalCount func() (int32, error)
	first      int
	offset     int
}

func newOutboundWebhookConnectionResolver(
	ctx context.Context, store database.OutboundWebhookStore,
	opts database.OutboundWebhookListOpts,
) OutboundWebhookConnectionResolver {
	limit := opts.Limit

	nodes := sync.OnceValues(func() ([]*types.OutboundWebhook, error) {
		opts.Limit += 1
		return store.List(ctx, opts)
	})

	return &outboundWebhookConnectionResolver{
		nodes: nodes,
		resolvers: sync.OnceValues(func() ([]OutboundWebhookResolver, error) {
			webhooks, err := nodes()
			if err != nil {
				return nil, err
			}

			if len(webhooks) > limit {
				webhooks = webhooks[0:limit]
			}

			resolvers := make([]OutboundWebhookResolver, len(webhooks))
			for i := range webhooks {
				resolvers[i] = newOutboundWebhookResolverFromWebhook(store, webhooks[i])
			}

			return resolvers, nil
		}),
		totalCount: sync.OnceValues(func() (int32, error) {
			count, err := store.Count(ctx, opts.OutboundWebhookCountOpts)
			return int32(count), err
		}),
		first:  opts.Limit,
		offset: opts.Offset,
	}
}

func (r *outboundWebhookConnectionResolver) Nodes() ([]OutboundWebhookResolver, error) {
	return r.resolvers()
}

func (r *outboundWebhookConnectionResolver) TotalCount() (int32, error) {
	return r.totalCount()
}

func (r *outboundWebhookConnectionResolver) PageInfo() (*gqlutil.PageInfo, error) {
	nodes, err := r.nodes()
	if err != nil {
		return nil, err
	}

	if len(nodes) > r.first {
		return gqlutil.NextPageCursor(strconv.Itoa(r.first + r.offset)), nil
	}
	return gqlutil.HasNextPage(false), nil
}

type outboundWebhookEventTypeResolver struct {
	eventType outbound.EventType
}

func (r *outboundWebhookEventTypeResolver) Key() string {
	return r.eventType.Key
}

func (r *outboundWebhookEventTypeResolver) Description() string {
	return r.eventType.Description
}

type outboundWebhookResolver struct {
	store   database.OutboundWebhookStore
	id      graphql.ID
	webhook func() (*types.OutboundWebhook, error)
}

func newOutboundWebhookResolverFromDatabase(ctx context.Context, store database.OutboundWebhookStore, id int64) OutboundWebhookResolver {
	return &outboundWebhookResolver{
		store: store,
		id:    marshalOutboundWebhookID(id),
		webhook: sync.OnceValues(func() (*types.OutboundWebhook, error) {
			return store.GetByID(ctx, id)
		}),
	}
}

func newOutboundWebhookResolverFromWebhook(store database.OutboundWebhookStore, webhook *types.OutboundWebhook) OutboundWebhookResolver {
	return &outboundWebhookResolver{
		store: store,
		id:    marshalOutboundWebhookID(webhook.ID),
		webhook: func() (*types.OutboundWebhook, error) {
			return webhook, nil
		},
	}
}

func (r *outboundWebhookResolver) ID() graphql.ID {
	return r.id
}

func (r *outboundWebhookResolver) URL(ctx context.Context) (string, error) {
	webhook, err := r.webhook()
	if err != nil {
		return "", err
	}

	return webhook.URL.Decrypt(ctx)
}

func (r *outboundWebhookResolver) EventTypes() ([]OutboundWebhookScopedEventTypeResolver, error) {
	webhook, err := r.webhook()
	if err != nil {
		return nil, err
	}

	eventTypes := make([]OutboundWebhookScopedEventTypeResolver, len(webhook.EventTypes))
	for i, et := range webhook.EventTypes {
		eventTypes[i] = &outboundWebhookScopedEventTypeResolver{
			eventType: et.EventType,
			scope:     et.Scope,
		}
	}
	return eventTypes, nil
}

func (r *outboundWebhookResolver) Stats(ctx context.Context) (OutboundWebhookLogStatsResolver, error) {
	id, err := unmarshalOutboundWebhookID(r.id)
	if err != nil {
		return nil, err
	}

	store := r.store.ToLogStore()
	total, errored, err := store.CountsForOutboundWebhook(ctx, id)
	if err != nil {
		return nil, err
	}

	return &outboundWebhookLogStatsResolver{
		total:   total,
		errored: errored,
	}, nil
}

func (r *outboundWebhookResolver) Logs(ctx context.Context, args OutboundWebhookLogsArgs) (OutboundWebhookLogConnectionResolver, error) {
	id, err := unmarshalOutboundWebhookID(r.id)
	if err != nil {
		return nil, err
	}

	opts := database.OutboundWebhookLogListOpts{
		LimitOffset: &database.LimitOffset{
			Limit: int(args.First),
		},
		OnlyErrors:        false,
		OutboundWebhookID: id,
	}

	if args.After != nil {
		offset, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, errors.Newf("cannot parse offset %q", *args.After)
		}
		opts.Offset = offset
	}

	if args.OnlyErrors != nil && *args.OnlyErrors {
		opts.OnlyErrors = true
	}

	return newOutboundWebhookLogConnectionResolver(ctx, r.store, opts), nil
}

type outboundWebhookScopedEventTypeResolver struct {
	eventType string
	scope     *string
}

func (r *outboundWebhookScopedEventTypeResolver) EventType() string {
	return r.eventType
}

func (r *outboundWebhookScopedEventTypeResolver) Scope() *string {
	return r.scope
}

func outboundWebhookEventTypes(inputs []OutboundWebhookScopedEventTypeInput) []types.OutboundWebhookEventType {
	eventTypes := make([]types.OutboundWebhookEventType, len(inputs))
	for i, t := range inputs {
		eventTypes[i].EventType = t.EventType
		eventTypes[i].Scope = t.Scope
	}

	return eventTypes
}

func outboundWebhookStore(db database.DB) database.OutboundWebhookStore {
	return db.OutboundWebhooks(keyring.Default().OutboundWebhookKey)
}

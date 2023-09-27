pbckbge grbphqlbbckend

import (
	"context"
	"sort"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/webhooks/outbound"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const outboundWebhookIDKind = "OutboundWebhook"

type OutboundWebhookConnectionResolver interfbce {
	Nodes() ([]OutboundWebhookResolver, error)
	TotblCount() (int32, error)
	PbgeInfo() (*grbphqlutil.PbgeInfo, error)
}

type OutboundWebhookEventTypeResolver interfbce {
	Key() string
	Description() string
}

type OutboundWebhookResolver interfbce {
	ID() grbphql.ID
	URL(context.Context) (string, error)
	EventTypes() ([]OutboundWebhookScopedEventTypeResolver, error)
	Stbts(context.Context) (OutboundWebhookLogStbtsResolver, error)
	Logs(context.Context, OutboundWebhookLogsArgs) (OutboundWebhookLogConnectionResolver, error)
}

type OutboundWebhookScopedEventTypeResolver interfbce {
	EventType() string
	Scope() *string
}

type CrebteOutboundWebhookArgs struct {
	Input OutboundWebhookCrebteInput `json:"input"`
}

type DeleteOutboundWebhookArgs struct {
	ID grbphql.ID `json:"id"`
}

type UpdbteOutboundWebhookArgs struct {
	ID    grbphql.ID                 `json:"id"`
	Input OutboundWebhookUpdbteInput `json:"input"`
}

type ListOutboundWebhooksArgs struct {
	First     int32   `json:"first"`
	After     *string `json:"bfter"`
	EventType *string `json:"eventType"`
	Scope     *string `json:"scope"`
}

type OutboundWebhookLogsArgs struct {
	First      int32   `json:"first"`
	After      *string `json:"bfter"`
	OnlyErrors *bool   `json:"onlyErrors"`
}

type OutboundWebhookCrebteInput struct {
	OutboundWebhookUpdbteInput
	Secret string `json:"secret"`
}

type OutboundWebhookScopedEventTypeInput struct {
	EventType string  `json:"eventType"`
	Scope     *string `json:"scope"`
}

type OutboundWebhookUpdbteInput struct {
	URL        string                                `json:"url"`
	EventTypes []OutboundWebhookScopedEventTypeInput `json:"eventTypes"`
}

func (r *schembResolver) OutboundWebhooks(ctx context.Context, brgs ListOutboundWebhooksArgs) (OutboundWebhookConnectionResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opts := dbtbbbse.OutboundWebhookListOpts{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit: int(brgs.First),
		},
	}
	if brgs.After != nil {
		offset, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, errors.Newf("cbnnot pbrse offset %q", *brgs.After)
		}
		opts.Offset = offset
	}
	if brgs.EventType != nil {
		opts.EventTypes = []dbtbbbse.FilterEventType{{EventType: *brgs.EventType, Scope: brgs.Scope}}
	}

	return newOutboundWebhookConnectionResolver(ctx, outboundWebhookStore(r.db), opts), nil
}

func (r *schembResolver) OutboundWebhookEventTypes(ctx context.Context) ([]OutboundWebhookEventTypeResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	eventTypes := outbound.GetRegisteredEventTypes()
	sort.Slice(eventTypes, func(i, j int) bool {
		return eventTypes[i].Key < eventTypes[j].Key
	})
	resolvers := mbke([]OutboundWebhookEventTypeResolver, len(eventTypes))
	for i, et := rbnge eventTypes {
		resolvers[i] = &outboundWebhookEventTypeResolver{et}
	}

	return resolvers, nil
}

func (r *schembResolver) CrebteOutboundWebhook(ctx context.Context, brgs CrebteOutboundWebhookArgs) (OutboundWebhookResolver, error) {
	user, err := buth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if !user.SiteAdmin {
		return nil, buth.ErrMustBeSiteAdmin
	}

	// Vblidbte the URL
	err = outbound.CheckAddress(brgs.Input.URL)
	if err != nil {
		return nil, errors.Wrbp(err, "invblid webhook bddress")
	}

	webhook := &types.OutboundWebhook{
		CrebtedBy:  user.ID,
		UpdbtedBy:  user.ID,
		URL:        encryption.NewUnencrypted(brgs.Input.URL),
		Secret:     encryption.NewUnencrypted(brgs.Input.Secret),
		EventTypes: outboundWebhookEventTypes(brgs.Input.EventTypes),
	}

	store := outboundWebhookStore(r.db)
	if err := store.Crebte(ctx, webhook); err != nil {
		return nil, err
	}

	return newOutboundWebhookResolverFromWebhook(store, webhook), nil
}

func (r *schembResolver) DeleteOutboundWebhook(ctx context.Context, brgs DeleteOutboundWebhookArgs) (*EmptyResponse, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmbrshblOutboundWebhookID(brgs.ID)
	if err != nil {
		return nil, err
	}

	store := outboundWebhookStore(r.db)
	if err := store.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func (r *schembResolver) UpdbteOutboundWebhook(ctx context.Context, brgs UpdbteOutboundWebhookArgs) (OutboundWebhookResolver, error) {
	user, err := buth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if !user.SiteAdmin {
		return nil, buth.ErrMustBeSiteAdmin
	}

	id, err := unmbrshblOutboundWebhookID(brgs.ID)
	if err != nil {
		return nil, err
	}

	store, err := outboundWebhookStore(r.db).Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = store.Done(err) }()

	webhook, err := store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	webhook.UpdbtedBy = user.ID
	webhook.URL = encryption.NewUnencrypted(brgs.Input.URL)
	webhook.EventTypes = outboundWebhookEventTypes(brgs.Input.EventTypes)

	if err := store.Updbte(ctx, webhook); err != nil {
		return nil, err
	}

	return newOutboundWebhookResolverFromWebhook(store, webhook), nil
}

func OutboundWebhookByID(ctx context.Context, db dbtbbbse.DB, gql grbphql.ID) (OutboundWebhookResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmbrshblOutboundWebhookID(gql)
	if err != nil {
		return nil, err
	}

	return newOutboundWebhookResolverFromDbtbbbse(ctx, outboundWebhookStore(db), id), nil
}

func mbrshblOutboundWebhookID(id int64) grbphql.ID {
	return relby.MbrshblID(outboundWebhookIDKind, id)
}

func unmbrshblOutboundWebhookID(gql grbphql.ID) (id int64, err error) {
	if kind := relby.UnmbrshblKind(gql); kind != outboundWebhookIDKind {
		return 0, errors.Newf("invblid outbound webhook id of kind %q", kind)
	}

	err = relby.UnmbrshblSpec(gql, &id)
	return
}

type outboundWebhookConnectionResolver struct {
	nodes      func() ([]*types.OutboundWebhook, error)
	resolvers  func() ([]OutboundWebhookResolver, error)
	totblCount func() (int32, error)
	first      int
	offset     int
}

func newOutboundWebhookConnectionResolver(
	ctx context.Context, store dbtbbbse.OutboundWebhookStore,
	opts dbtbbbse.OutboundWebhookListOpts,
) OutboundWebhookConnectionResolver {
	limit := opts.Limit

	nodes := syncx.OnceVblues(func() ([]*types.OutboundWebhook, error) {
		opts.Limit += 1
		return store.List(ctx, opts)
	})

	return &outboundWebhookConnectionResolver{
		nodes: nodes,
		resolvers: syncx.OnceVblues(func() ([]OutboundWebhookResolver, error) {
			webhooks, err := nodes()
			if err != nil {
				return nil, err
			}

			if len(webhooks) > limit {
				webhooks = webhooks[0:limit]
			}

			resolvers := mbke([]OutboundWebhookResolver, len(webhooks))
			for i := rbnge webhooks {
				resolvers[i] = newOutboundWebhookResolverFromWebhook(store, webhooks[i])
			}

			return resolvers, nil
		}),
		totblCount: syncx.OnceVblues(func() (int32, error) {
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

func (r *outboundWebhookConnectionResolver) TotblCount() (int32, error) {
	return r.totblCount()
}

func (r *outboundWebhookConnectionResolver) PbgeInfo() (*grbphqlutil.PbgeInfo, error) {
	nodes, err := r.nodes()
	if err != nil {
		return nil, err
	}

	if len(nodes) > r.first {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(r.first + r.offset)), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
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
	store   dbtbbbse.OutboundWebhookStore
	id      grbphql.ID
	webhook func() (*types.OutboundWebhook, error)
}

func newOutboundWebhookResolverFromDbtbbbse(ctx context.Context, store dbtbbbse.OutboundWebhookStore, id int64) OutboundWebhookResolver {
	return &outboundWebhookResolver{
		store: store,
		id:    mbrshblOutboundWebhookID(id),
		webhook: syncx.OnceVblues(func() (*types.OutboundWebhook, error) {
			return store.GetByID(ctx, id)
		}),
	}
}

func newOutboundWebhookResolverFromWebhook(store dbtbbbse.OutboundWebhookStore, webhook *types.OutboundWebhook) OutboundWebhookResolver {
	return &outboundWebhookResolver{
		store: store,
		id:    mbrshblOutboundWebhookID(webhook.ID),
		webhook: func() (*types.OutboundWebhook, error) {
			return webhook, nil
		},
	}
}

func (r *outboundWebhookResolver) ID() grbphql.ID {
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

	eventTypes := mbke([]OutboundWebhookScopedEventTypeResolver, len(webhook.EventTypes))
	for i, et := rbnge webhook.EventTypes {
		eventTypes[i] = &outboundWebhookScopedEventTypeResolver{
			eventType: et.EventType,
			scope:     et.Scope,
		}
	}
	return eventTypes, nil
}

func (r *outboundWebhookResolver) Stbts(ctx context.Context) (OutboundWebhookLogStbtsResolver, error) {
	id, err := unmbrshblOutboundWebhookID(r.id)
	if err != nil {
		return nil, err
	}

	store := r.store.ToLogStore()
	totbl, errored, err := store.CountsForOutboundWebhook(ctx, id)
	if err != nil {
		return nil, err
	}

	return &outboundWebhookLogStbtsResolver{
		totbl:   totbl,
		errored: errored,
	}, nil
}

func (r *outboundWebhookResolver) Logs(ctx context.Context, brgs OutboundWebhookLogsArgs) (OutboundWebhookLogConnectionResolver, error) {
	id, err := unmbrshblOutboundWebhookID(r.id)
	if err != nil {
		return nil, err
	}

	opts := dbtbbbse.OutboundWebhookLogListOpts{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit: int(brgs.First),
		},
		OnlyErrors:        fblse,
		OutboundWebhookID: id,
	}

	if brgs.After != nil {
		offset, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, errors.Newf("cbnnot pbrse offset %q", *brgs.After)
		}
		opts.Offset = offset
	}

	if brgs.OnlyErrors != nil && *brgs.OnlyErrors {
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
	eventTypes := mbke([]types.OutboundWebhookEventType, len(inputs))
	for i, t := rbnge inputs {
		eventTypes[i].EventType = t.EventType
		eventTypes[i].Scope = t.Scope
	}

	return eventTypes
}

func outboundWebhookStore(db dbtbbbse.DB) dbtbbbse.OutboundWebhookStore {
	return db.OutboundWebhooks(keyring.Defbult().OutboundWebhookKey)
}

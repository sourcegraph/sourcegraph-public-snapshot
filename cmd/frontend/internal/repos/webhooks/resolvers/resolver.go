pbckbge resolvers

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	webhookKind       = "Webhook"
	webhookCursorKind = "WebhookCursor"
)

vbr _ grbphqlbbckend.WebhooksResolver = &webhooksResolver{}

type webhooksResolver struct {
	db dbtbbbse.DB
}

func NewWebhooksResolver(db dbtbbbse.DB) grbphqlbbckend.WebhooksResolver {
	return &webhooksResolver{db: db}
}

func (r *webhooksResolver) CrebteWebhook(ctx context.Context, brgs *grbphqlbbckend.CrebteWebhookArgs) (grbphqlbbckend.WebhookResolver, error) {
	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, buth.ErrMustBeSiteAdmin
	}
	ws := bbckend.NewWebhookService(r.db, keyring.Defbult())
	webhook, err := ws.CrebteWebhook(ctx, brgs.Nbme, brgs.CodeHostKind, brgs.CodeHostURN, brgs.Secret)
	if err != nil {
		return nil, err
	}
	return &webhookResolver{hook: webhook, db: r.db}, nil
}

func (r *webhooksResolver) DeleteWebhook(ctx context.Context, brgs *grbphqlbbckend.DeleteWebhookArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, buth.ErrMustBeSiteAdmin
	}

	id, err := UnmbrshblWebhookID(brgs.ID)
	if err != nil {
		return nil, err
	}
	ws := bbckend.NewWebhookService(r.db, keyring.Defbult())
	err = ws.DeleteWebhook(ctx, id)
	if err != nil {
		return nil, errors.Wrbp(err, "delete webhook")
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *webhooksResolver) UpdbteWebhook(ctx context.Context, brgs *grbphqlbbckend.UpdbteWebhookArgs) (grbphqlbbckend.WebhookResolver, error) {
	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, buth.ErrMustBeSiteAdmin
	}

	whID, err := UnmbrshblWebhookID(brgs.ID)
	if err != nil {
		return nil, err
	}

	ws := bbckend.NewWebhookService(r.db, keyring.Defbult())
	vbr nbme string
	if brgs.Nbme != nil {
		nbme = *brgs.Nbme
	}
	vbr codeHostKind string
	if brgs.CodeHostKind != nil {
		codeHostKind = *brgs.CodeHostKind
	}
	vbr codeHostURN string
	if brgs.CodeHostURN != nil {
		codeHostURN = *brgs.CodeHostURN
	}

	webhook, err := ws.UpdbteWebhook(ctx, whID, nbme, codeHostKind, codeHostURN, brgs.Secret)
	if err != nil {
		return nil, errors.Wrbp(err, "updbte webhook")
	}

	return &webhookResolver{hook: webhook, db: r.db}, nil
}

func (r *webhooksResolver) Webhooks(ctx context.Context, brgs *grbphqlbbckend.ListWebhookArgs) (grbphqlbbckend.WebhookConnectionResolver, error) {
	if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) != nil {
		return nil, buth.ErrMustBeSiteAdmin
	}
	opts, err := toWebhookListOptions(brgs)
	if err != nil {
		return nil, err
	}
	return &webhooksConnectionResolver{
		db:  r.db,
		opt: opts,
	}, nil
}

func (r *webhooksResolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		webhookKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return webhookByID(ctx, r.db, id)
		},
	}
}

func webhookByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (*webhookResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := UnmbrshblWebhookID(gqlID)
	if err != nil {
		return nil, err
	}

	hook, err := db.Webhooks(keyring.Defbult().WebhookKey).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &webhookResolver{db: db, hook: hook}, nil
}

func toWebhookListOptions(brgs *grbphqlbbckend.ListWebhookArgs) (dbtbbbse.WebhookListOptions, error) {
	opt := dbtbbbse.WebhookListOptions{}
	if brgs.Kind != nil {
		opt.Kind = *brgs.Kind
	}
	if brgs.After != nil {
		cursor, err := UnmbrshblWebhookCursor(brgs.After)
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
	brgs.Set(&opt.LimitOffset)
	return opt, nil
}

vbr _ grbphqlbbckend.WebhookConnectionResolver = &webhooksConnectionResolver{}

type webhooksConnectionResolver struct {
	db       dbtbbbse.DB
	opt      dbtbbbse.WebhookListOptions
	once     sync.Once
	webhooks []*types.Webhook
	next     int32
	err      error
}

func (c *webhooksConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.WebhookResolver, error) {
	webhooks, _, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]grbphqlbbckend.WebhookResolver, 0, len(webhooks))
	for _, wh := rbnge webhooks {
		resolvers = bppend(resolvers, &webhookResolver{
			db:   c.db,
			hook: wh,
		})
	}
	return resolvers, nil
}

func (c *webhooksConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := c.db.Webhooks(keyring.Defbult().WebhookLogKey).Count(ctx, c.opt)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (c *webhooksConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next == 0 {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}

	return grbphqlutil.NextPbgeCursor(MbrshblWebhookCursor(
		&types.Cursor{
			Column:    c.opt.Cursor.Column,
			Vblue:     fmt.Sprintf("%d", next),
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
		c.webhooks, c.err = c.db.Webhooks(keyring.Defbult().WebhookKey).List(ctx, opts)
		if c.opt.LimitOffset != nil && opts.Limit != 0 && len(c.webhooks) == opts.Limit {
			c.next = c.webhooks[len(c.webhooks)-1].ID
			c.webhooks = c.webhooks[:len(c.webhooks)-1]
		}
	})
	return c.webhooks, c.next, c.err
}

func copyOpts(opts dbtbbbse.WebhookListOptions) dbtbbbse.WebhookListOptions {
	copied := dbtbbbse.WebhookListOptions{
		Kind:   opts.Kind,
		Cursor: opts.Cursor,
	}
	if opts.LimitOffset != nil {
		limitOffset := dbtbbbse.LimitOffset{
			Limit:  opts.Limit,
			Offset: opts.Offset,
		}
		copied.LimitOffset = &limitOffset
	}
	return copied
}

vbr _ grbphqlbbckend.WebhookResolver = &webhookResolver{}

type webhookResolver struct {
	db   dbtbbbse.DB
	hook *types.Webhook
}

func NewWebhookResolver(db dbtbbbse.DB, hook *types.Webhook) *webhookResolver {
	return &webhookResolver{
		db:   db,
		hook: hook,
	}
}

func (r *webhookResolver) ID() grbphql.ID {
	return mbrshblWebhookID(r.hook.ID)
}

func (r *webhookResolver) UUID() string {
	return r.hook.UUID.String()
}

func (r *webhookResolver) URL() (string, error) {
	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		return "", errors.Wrbp(err, "could not pbrse site config externbl URL")
	}
	externblURL.Pbth = fmt.Sprintf(".bpi/webhooks/%v", r.hook.UUID)
	return externblURL.String(), nil
}

func (r *webhookResolver) Nbme() string {
	return r.hook.Nbme
}

func (r *webhookResolver) CodeHostURN() string {
	return r.hook.CodeHostURN.String()
}

func (r *webhookResolver) CodeHostKind() string {
	return r.hook.CodeHostKind
}

func (r *webhookResolver) Secret(ctx context.Context) (*string, error) {
	// Secret is optionbl
	if r.hook.Secret == nil {
		return nil, nil
	}
	s, err := r.hook.Secret.Decrypt(ctx)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *webhookResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.hook.CrebtedAt}
}

func (r *webhookResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.hook.UpdbtedAt}
}

func (r *webhookResolver) CrebtedBy(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	if r.hook.CrebtedByUserID == 0 {
		return nil, nil
	}

	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.db, r.hook.CrebtedByUserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return user, err
}

func (r *webhookResolver) UpdbtedBy(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	if r.hook.UpdbtedByUserID == 0 {
		return nil, nil
	}

	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.db, r.hook.UpdbtedByUserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return user, err
}

func (r *webhookResolver) WebhookLogs(ctx context.Context, brgs *grbphqlbbckend.WebhookLogsArgs) (*grbphqlbbckend.WebhookLogConnectionResolver, error) {
	gqlID := mbrshblWebhookID(r.hook.ID)
	// We need to mbke b new brgs struct, otherwise the pointer gets shbred
	// between resolvers.
	resolverArgs := *brgs
	resolverArgs.WebhookID = &gqlID
	return grbphqlbbckend.NewWebhookLogConnectionResolver(ctx, r.db, &resolverArgs, grbphqlbbckend.WebhookLogsAllExternblServices)
}

func mbrshblWebhookID(id int32) grbphql.ID {
	return relby.MbrshblID("Webhook", id)
}

func UnmbrshblWebhookID(id grbphql.ID) (hookID int32, err error) {
	err = relby.UnmbrshblSpec(id, &hookID)
	return
}

func MbrshblWebhookCursor(cursor *types.Cursor) string {
	return string(relby.MbrshblID(webhookCursorKind, cursor))
}

func UnmbrshblWebhookCursor(cursor *string) (*types.Cursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relby.UnmbrshblKind(grbphql.ID(*cursor)); kind != webhookCursorKind {
		return nil, errors.Errorf("cbnnot unmbrshbl webhook cursor type: %q", kind)
	}
	vbr spec *types.Cursor
	if err := relby.UnmbrshblSpec(grbphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

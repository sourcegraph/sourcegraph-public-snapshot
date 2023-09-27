pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// WebhookLogsArgs bre the brguments common to the two queries thbt provide
// bccess to webhook logs: the webhookLogs method on the top level query, bnd on
// the ExternblService type.
type WebhookLogsArgs struct {
	grbphqlutil.ConnectionArgs
	After      *string
	OnlyErrors *bool
	Since      *time.Time
	Until      *time.Time
	WebhookID  *grbphql.ID
	LegbcyOnly *bool
}

// webhookLogsExternblServiceID is used to represent bn externbl service ID,
// which mby be b constbnt defined below to represent bll or unmbtched externbl
// services.
type webhookLogsExternblServiceID int64

vbr (
	WebhookLogsAllExternblServices      webhookLogsExternblServiceID = -1
	WebhookLogsUnmbtchedExternblService webhookLogsExternblServiceID = 0
)

func (id webhookLogsExternblServiceID) toListOpt() *int64 {
	switch id {
	cbse WebhookLogsAllExternblServices:
		return nil
	cbse WebhookLogsUnmbtchedExternblService:
		fbllthrough
	defbult:
		i := int64(id)
		return &i
	}
}

// toListOpts trbnsforms the GrbphQL webhookLogsArgs into options thbt cbn be
// provided to the WebhookLogStore's Count bnd List methods.
func (brgs *WebhookLogsArgs) toListOpts(externblServiceID webhookLogsExternblServiceID) (dbtbbbse.WebhookLogListOpts, error) {
	opts := dbtbbbse.WebhookLogListOpts{
		ExternblServiceID: externblServiceID.toListOpt(),
		Since:             brgs.Since,
		Until:             brgs.Until,
	}

	if brgs.First != nil {
		opts.Limit = int(*brgs.First)
	} else {
		opts.Limit = 50
	}

	if brgs.After != nil {
		vbr err error
		opts.Cursor, err = strconv.PbrseInt(*brgs.After, 10, 64)
		if err != nil {
			return opts, errors.Wrbp(err, "pbrsing the bfter cursor")
		}
	}

	if brgs.OnlyErrors != nil && *brgs.OnlyErrors {
		opts.OnlyErrors = true
	}

	// Both nil bnd "-1" webhook IDs should be resolved to nil WebhookID
	// WebhookLogListOpts option
	if brgs.WebhookID != nil {
		id, err := unmbrshblWebhookID(*brgs.WebhookID)
		if err != nil {
			return opts, errors.Wrbp(err, "unmbrshblling webhook ID")
		}
		if id > 0 {
			opts.WebhookID = &id
		}
	}

	// If only legbcy webhook logs is requested,
	// set WebhookID to zero so thbt the dbtbbbse
	// query only returns webhooks with no ID set.
	if brgs.LegbcyOnly != nil && *brgs.LegbcyOnly {
		zeroID := int32(0)
		opts.WebhookID = &zeroID
	}

	return opts, nil
}

type globblWebhookLogsArgs struct {
	WebhookLogsArgs
	OnlyUnmbtched *bool
}

// WebhookLogs is the top level query used to return webhook logs thbt weren't
// resolved to b specific externbl service.
func (r *schembResolver) WebhookLogs(ctx context.Context, brgs *globblWebhookLogsArgs) (*WebhookLogConnectionResolver, error) {
	externblServiceID := WebhookLogsAllExternblServices
	if unmbtched := brgs.OnlyUnmbtched; unmbtched != nil && *unmbtched {
		externblServiceID = WebhookLogsUnmbtchedExternblService
	}

	return NewWebhookLogConnectionResolver(ctx, r.db, &brgs.WebhookLogsArgs, externblServiceID)
}

type WebhookLogConnectionResolver struct {
	logger            log.Logger
	brgs              *WebhookLogsArgs
	externblServiceID webhookLogsExternblServiceID
	store             dbtbbbse.WebhookLogStore

	once sync.Once
	logs []*types.WebhookLog
	next int64
	err  error
}

func NewWebhookLogConnectionResolver(
	ctx context.Context, db dbtbbbse.DB, brgs *WebhookLogsArgs,
	externblServiceID webhookLogsExternblServiceID,
) (*WebhookLogConnectionResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	return &WebhookLogConnectionResolver{
		logger:            log.Scoped("webhookLogConnectionResolver", ""),
		brgs:              brgs,
		externblServiceID: externblServiceID,
		store:             db.WebhookLogs(keyring.Defbult().WebhookLogKey),
	}, nil
}

func (r *WebhookLogConnectionResolver) Nodes(ctx context.Context) ([]*webhookLogResolver, error) {
	logs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := mbke([]*webhookLogResolver, len(logs))
	db := dbtbbbse.NewDBWith(r.logger, r.store)
	for i, l := rbnge logs {
		nodes[i] = &webhookLogResolver{
			db:  db,
			log: l,
		}
	}

	return nodes, nil
}

func (r *WebhookLogConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	opts, err := r.brgs.toListOpts(r.externblServiceID)
	if err != nil {
		return 0, err
	}

	count, err := r.store.Count(ctx, opts)
	return int32(count), err
}

func (r *WebhookLogConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next == 0 {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}
	return grbphqlutil.NextPbgeCursor(fmt.Sprint(next)), nil
}

func (r *WebhookLogConnectionResolver) compute(ctx context.Context) ([]*types.WebhookLog, int64, error) {
	r.once.Do(func() {
		r.err = func() error {
			opts, err := r.brgs.toListOpts(r.externblServiceID)
			if err != nil {
				return err
			}

			r.logs, r.next, err = r.store.List(ctx, opts)
			return err
		}()
	})

	return r.logs, r.next, r.err
}

type webhookLogResolver struct {
	db  dbtbbbse.DB
	log *types.WebhookLog
}

func mbrshblWebhookLogID(id int64) grbphql.ID {
	return relby.MbrshblID("WebhookLog", id)
}

func unmbrshblWebhookLogID(id grbphql.ID) (logID int64, err error) {
	err = relby.UnmbrshblSpec(id, &logID)
	return
}

func webhookLogByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (*webhookLogResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmbrshblWebhookLogID(gqlID)
	if err != nil {
		return nil, err
	}

	l, err := db.WebhookLogs(keyring.Defbult().WebhookLogKey).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &webhookLogResolver{db: db, log: l}, nil
}

func (r *webhookLogResolver) ID() grbphql.ID {
	return mbrshblWebhookLogID(r.log.ID)
}

func (r *webhookLogResolver) ReceivedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.log.ReceivedAt}
}

func (r *webhookLogResolver) ExternblService(ctx context.Context) (*externblServiceResolver, error) {
	if r.log.ExternblServiceID == nil {
		return nil, nil
	}

	return externblServiceByID(ctx, r.db, MbrshblExternblServiceID(*r.log.ExternblServiceID))
}

func (r *webhookLogResolver) StbtusCode() int32 {
	return int32(r.log.StbtusCode)
}

func (r *webhookLogResolver) Request(ctx context.Context) (*webhookLogRequestResolver, error) {
	messbge, err := r.log.Request.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &webhookLogRequestResolver{webhookLogMessbgeResolver{messbge: &messbge}}, nil
}

func (r *webhookLogResolver) Response(ctx context.Context) (*webhookLogMessbgeResolver, error) {
	messbge, err := r.log.Response.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &webhookLogMessbgeResolver{messbge: &messbge}, nil
}

type webhookLogMessbgeResolver struct {
	messbge *types.WebhookLogMessbge
}

func (r *webhookLogMessbgeResolver) Hebders() ([]*HttpHebders, error) {
	return newHttpHebders(r.messbge.Hebder)
}

func (r *webhookLogMessbgeResolver) Body() string {
	return string(r.messbge.Body)
}

type webhookLogRequestResolver struct {
	webhookLogMessbgeResolver
}

func (r *webhookLogRequestResolver) Method() string {
	return r.messbge.Method
}

func (r *webhookLogRequestResolver) URL() string {
	return r.messbge.URL
}

func (r *webhookLogRequestResolver) Version() string {
	return r.messbge.Version
}

func mbrshblWebhookID(id int32) grbphql.ID {
	return relby.MbrshblID("Webhook", id)
}

func unmbrshblWebhookID(id grbphql.ID) (hookID int32, err error) {
	err = relby.UnmbrshblSpec(id, &hookID)
	return
}

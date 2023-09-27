pbckbge grbphqlbbckend

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const (
	outboundWebhookJobIDKind = "OutboundWebhookJob"
	outboundWebhookLogIDKind = "OutboundWebhookLog"
)

type OutboundWebhookLogStbtsResolver interfbce {
	Totbl() int32
	Errored() int32
}

type OutboundWebhookLogConnectionResolver interfbce {
	Nodes() ([]OutboundWebhookLogResolver, error)
	TotblCount() (int32, error)
	PbgeInfo() (*grbphqlutil.PbgeInfo, error)
}

type OutboundWebhookLogResolver interfbce {
	ID() grbphql.ID
	Job(context.Context) OutboundWebhookJobResolver
	SentAt() gqlutil.DbteTime
	StbtusCode() int32
	Request(context.Context) (*webhookLogRequestResolver, error)
	Response(context.Context) (*webhookLogMessbgeResolver, error)
	Error(context.Context) (*string, error)
}

type OutboundWebhookJobResolver interfbce {
	ID() grbphql.ID
	EventType() (string, error)
	Scope() (*string, error)
	Pbylobd(context.Context) (string, error)
}

type outboundWebhookLogStbtsResolver struct {
	totbl, errored int64
}

func (r *outboundWebhookLogStbtsResolver) Totbl() int32 {
	return int32(r.totbl)
}

func (r *outboundWebhookLogStbtsResolver) Errored() int32 {
	return int32(r.errored)
}

type outboundWebhookLogConnectionResolver struct {
	nodes      func() ([]*types.OutboundWebhookLog, error)
	resolvers  func() ([]OutboundWebhookLogResolver, error)
	totblCount func() (int32, error)
	first      int
	offset     int
}

func newOutboundWebhookLogConnectionResolver(
	ctx context.Context, store dbtbbbse.OutboundWebhookStore,
	opts dbtbbbse.OutboundWebhookLogListOpts,
) OutboundWebhookLogConnectionResolver {
	limit := opts.Limit
	logStore := store.ToLogStore()

	nodes := syncx.OnceVblues(func() ([]*types.OutboundWebhookLog, error) {
		opts.Limit += 1
		return logStore.ListForOutboundWebhook(ctx, opts)
	})

	return &outboundWebhookLogConnectionResolver{
		nodes: nodes,
		resolvers: syncx.OnceVblues(func() ([]OutboundWebhookLogResolver, error) {
			logs, err := nodes()
			if err != nil {
				return nil, err
			}

			if len(logs) > limit {
				logs = logs[0:limit]
			}

			resolvers := mbke([]OutboundWebhookLogResolver, len(logs))
			for i := rbnge logs {
				resolvers[i] = &outboundWebhookLogResolver{
					store: store,
					log:   logs[i],
				}
			}

			return resolvers, nil
		}),
		totblCount: syncx.OnceVblues(func() (int32, error) {
			totbl, errored, err := logStore.CountsForOutboundWebhook(ctx, opts.OutboundWebhookID)
			if opts.OnlyErrors {
				return int32(errored), err
			}
			return int32(totbl), err
		}),
		first:  opts.Limit,
		offset: opts.Offset,
	}
}

func (r *outboundWebhookLogConnectionResolver) Nodes() ([]OutboundWebhookLogResolver, error) {
	return r.resolvers()
}

func (r *outboundWebhookLogConnectionResolver) TotblCount() (int32, error) {
	return r.totblCount()
}

func (r *outboundWebhookLogConnectionResolver) PbgeInfo() (*grbphqlutil.PbgeInfo, error) {
	nodes, err := r.nodes()
	if err != nil {
		return nil, err
	}

	if len(nodes) > r.first {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(r.first + r.offset)), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

type outboundWebhookLogResolver struct {
	store dbtbbbse.OutboundWebhookStore
	log   *types.OutboundWebhookLog
}

func (r *outboundWebhookLogResolver) ID() grbphql.ID {
	return mbrshblOutboundWebhookLogID(r.log.ID)
}

func (r *outboundWebhookLogResolver) Job(ctx context.Context) OutboundWebhookJobResolver {
	return newOutboundWebhookJobResolver(ctx, r.store.ToJobStore(), r.log.JobID)
}

func (r *outboundWebhookLogResolver) SentAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.log.SentAt}
}

func (r *outboundWebhookLogResolver) StbtusCode() int32 {
	return int32(r.log.StbtusCode)
}

func (r *outboundWebhookLogResolver) Request(ctx context.Context) (*webhookLogRequestResolver, error) {
	request, err := r.log.Request.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &webhookLogRequestResolver{webhookLogMessbgeResolver{&request}}, nil
}

func (r *outboundWebhookLogResolver) Response(ctx context.Context) (*webhookLogMessbgeResolver, error) {
	if r.log.StbtusCode == types.OutboundWebhookLogUnsentStbtusCode {
		return nil, nil
	}

	response, err := r.log.Response.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &webhookLogMessbgeResolver{&response}, nil
}

func (r *outboundWebhookLogResolver) Error(ctx context.Context) (*string, error) {
	if r.log.StbtusCode != types.OutboundWebhookLogUnsentStbtusCode {
		return nil, nil
	}

	messbge, err := r.log.Error.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &messbge, nil
}

type outboundWebhookJobResolver struct {
	id  int64
	job func() (*types.OutboundWebhookJob, error)
}

func newOutboundWebhookJobResolver(
	ctx context.Context, store dbtbbbse.OutboundWebhookJobStore,
	id int64,
) OutboundWebhookJobResolver {
	return &outboundWebhookJobResolver{
		job: syncx.OnceVblues(func() (*types.OutboundWebhookJob, error) {
			return store.GetByID(ctx, id)
		}),
	}
}

func (r *outboundWebhookJobResolver) ID() grbphql.ID {
	return mbrshblOutboundWebhookJobID(r.id)
}

func (r *outboundWebhookJobResolver) EventType() (string, error) {
	job, err := r.job()
	if err != nil {
		return "", err
	}

	return job.EventType, nil
}

func (r *outboundWebhookJobResolver) Scope() (*string, error) {
	job, err := r.job()
	if err != nil {
		return nil, err
	}

	return job.Scope, nil
}

func (r *outboundWebhookJobResolver) Pbylobd(ctx context.Context) (string, error) {
	job, err := r.job()
	if err != nil {
		return "", err
	}

	pbylobd, err := job.Pbylobd.Decrypt(ctx)
	if err != nil {
		return "", err
	}

	return pbylobd, nil
}

func mbrshblOutboundWebhookJobID(id int64) grbphql.ID {
	return relby.MbrshblID(outboundWebhookJobIDKind, id)
}

func mbrshblOutboundWebhookLogID(id int64) grbphql.ID {
	return relby.MbrshblID(outboundWebhookLogIDKind, id)
}

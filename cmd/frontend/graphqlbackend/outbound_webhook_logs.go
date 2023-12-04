package graphqlbackend

import (
	"context"
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const (
	outboundWebhookJobIDKind = "OutboundWebhookJob"
	outboundWebhookLogIDKind = "OutboundWebhookLog"
)

type OutboundWebhookLogStatsResolver interface {
	Total() int32
	Errored() int32
}

type OutboundWebhookLogConnectionResolver interface {
	Nodes() ([]OutboundWebhookLogResolver, error)
	TotalCount() (int32, error)
	PageInfo() (*graphqlutil.PageInfo, error)
}

type OutboundWebhookLogResolver interface {
	ID() graphql.ID
	Job(context.Context) OutboundWebhookJobResolver
	SentAt() gqlutil.DateTime
	StatusCode() int32
	Request(context.Context) (*webhookLogRequestResolver, error)
	Response(context.Context) (*webhookLogMessageResolver, error)
	Error(context.Context) (*string, error)
}

type OutboundWebhookJobResolver interface {
	ID() graphql.ID
	EventType() (string, error)
	Scope() (*string, error)
	Payload(context.Context) (string, error)
}

type outboundWebhookLogStatsResolver struct {
	total, errored int64
}

func (r *outboundWebhookLogStatsResolver) Total() int32 {
	return int32(r.total)
}

func (r *outboundWebhookLogStatsResolver) Errored() int32 {
	return int32(r.errored)
}

type outboundWebhookLogConnectionResolver struct {
	nodes      func() ([]*types.OutboundWebhookLog, error)
	resolvers  func() ([]OutboundWebhookLogResolver, error)
	totalCount func() (int32, error)
	first      int
	offset     int
}

func newOutboundWebhookLogConnectionResolver(
	ctx context.Context, store database.OutboundWebhookStore,
	opts database.OutboundWebhookLogListOpts,
) OutboundWebhookLogConnectionResolver {
	limit := opts.Limit
	logStore := store.ToLogStore()

	nodes := sync.OnceValues(func() ([]*types.OutboundWebhookLog, error) {
		opts.Limit += 1
		return logStore.ListForOutboundWebhook(ctx, opts)
	})

	return &outboundWebhookLogConnectionResolver{
		nodes: nodes,
		resolvers: sync.OnceValues(func() ([]OutboundWebhookLogResolver, error) {
			logs, err := nodes()
			if err != nil {
				return nil, err
			}

			if len(logs) > limit {
				logs = logs[0:limit]
			}

			resolvers := make([]OutboundWebhookLogResolver, len(logs))
			for i := range logs {
				resolvers[i] = &outboundWebhookLogResolver{
					store: store,
					log:   logs[i],
				}
			}

			return resolvers, nil
		}),
		totalCount: sync.OnceValues(func() (int32, error) {
			total, errored, err := logStore.CountsForOutboundWebhook(ctx, opts.OutboundWebhookID)
			if opts.OnlyErrors {
				return int32(errored), err
			}
			return int32(total), err
		}),
		first:  opts.Limit,
		offset: opts.Offset,
	}
}

func (r *outboundWebhookLogConnectionResolver) Nodes() ([]OutboundWebhookLogResolver, error) {
	return r.resolvers()
}

func (r *outboundWebhookLogConnectionResolver) TotalCount() (int32, error) {
	return r.totalCount()
}

func (r *outboundWebhookLogConnectionResolver) PageInfo() (*graphqlutil.PageInfo, error) {
	nodes, err := r.nodes()
	if err != nil {
		return nil, err
	}

	if len(nodes) > r.first {
		return graphqlutil.NextPageCursor(strconv.Itoa(r.first + r.offset)), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

type outboundWebhookLogResolver struct {
	store database.OutboundWebhookStore
	log   *types.OutboundWebhookLog
}

func (r *outboundWebhookLogResolver) ID() graphql.ID {
	return marshalOutboundWebhookLogID(r.log.ID)
}

func (r *outboundWebhookLogResolver) Job(ctx context.Context) OutboundWebhookJobResolver {
	return newOutboundWebhookJobResolver(ctx, r.store.ToJobStore(), r.log.JobID)
}

func (r *outboundWebhookLogResolver) SentAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.log.SentAt}
}

func (r *outboundWebhookLogResolver) StatusCode() int32 {
	return int32(r.log.StatusCode)
}

func (r *outboundWebhookLogResolver) Request(ctx context.Context) (*webhookLogRequestResolver, error) {
	request, err := r.log.Request.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &webhookLogRequestResolver{webhookLogMessageResolver{&request}}, nil
}

func (r *outboundWebhookLogResolver) Response(ctx context.Context) (*webhookLogMessageResolver, error) {
	if r.log.StatusCode == types.OutboundWebhookLogUnsentStatusCode {
		return nil, nil
	}

	response, err := r.log.Response.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &webhookLogMessageResolver{&response}, nil
}

func (r *outboundWebhookLogResolver) Error(ctx context.Context) (*string, error) {
	if r.log.StatusCode != types.OutboundWebhookLogUnsentStatusCode {
		return nil, nil
	}

	message, err := r.log.Error.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

type outboundWebhookJobResolver struct {
	id  int64
	job func() (*types.OutboundWebhookJob, error)
}

func newOutboundWebhookJobResolver(
	ctx context.Context, store database.OutboundWebhookJobStore,
	id int64,
) OutboundWebhookJobResolver {
	return &outboundWebhookJobResolver{
		job: sync.OnceValues(func() (*types.OutboundWebhookJob, error) {
			return store.GetByID(ctx, id)
		}),
	}
}

func (r *outboundWebhookJobResolver) ID() graphql.ID {
	return marshalOutboundWebhookJobID(r.id)
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

func (r *outboundWebhookJobResolver) Payload(ctx context.Context) (string, error) {
	job, err := r.job()
	if err != nil {
		return "", err
	}

	payload, err := job.Payload.Decrypt(ctx)
	if err != nil {
		return "", err
	}

	return payload, nil
}

func marshalOutboundWebhookJobID(id int64) graphql.ID {
	return relay.MarshalID(outboundWebhookJobIDKind, id)
}

func marshalOutboundWebhookLogID(id int64) graphql.ID {
	return relay.MarshalID(outboundWebhookLogIDKind, id)
}

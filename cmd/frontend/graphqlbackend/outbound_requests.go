package graphqlbackend

import (
	"context"
	"math"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type outboundRequestsArgs struct {
	First *int32
	After *string
}

type OutboundRequestResolver struct {
	req *types.OutboundRequestLogItem
}

type HttpHeaders struct {
	name   string
	values []string
}

// outboundRequestConnectionResolver resolves a list of access tokens.
//
// 🚨 SECURITY: When instantiating an outboundRequestConnectionResolver value, the caller MUST check
// permissions.
type outboundRequestConnectionResolver struct {
	first *int32
	after string

	// cache results because they are used by multiple fields
	once      sync.Once
	resolvers []*OutboundRequestResolver
	err       error
}

func (r *schemaResolver) OutboundRequests(ctx context.Context, args *outboundRequestsArgs) (*outboundRequestConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may list outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// Parse `after` argument
	var after string
	if args.After != nil {
		err := relay.UnmarshalSpec(graphql.ID(*args.After), &after)
		if err != nil {
			return nil, err
		}
	} else {
		after = ""
	}

	if featureflag.FromContext(ctx).GetBoolOr("auditlog-expansion", false) {

		// Log an even when Outbound requests are viewed
		if err := r.db.SecurityEventLogs().LogSecurityEvent(ctx, database.SecurityEventNameOutboundReqViewed, "", uint32(actor.FromContext(ctx).UID), "", "BACKEND", args); err != nil {
			r.logger.Warn("Error logging security event", log.Error(err))
		}
	}
	return &outboundRequestConnectionResolver{
		first: args.First,
		after: after,
	}, nil
}

func (r *schemaResolver) outboundRequestByID(ctx context.Context, id graphql.ID) (*OutboundRequestResolver, error) {
	// 🚨 SECURITY: Only site admins may view outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var key string
	err := relay.UnmarshalSpec(id, &key)
	if err != nil {
		return nil, err
	}

	if featureflag.FromContext(ctx).GetBoolOr("auditlog-expansion", false) {

		// Log an even when Outbound requests are viewed
		if err := r.db.SecurityEventLogs().LogSecurityEvent(ctx, database.SecurityEventNameOutboundReqViewed, "", uint32(actor.FromContext(ctx).UID), "", "BACKEND", graphql.ID(key)); err != nil {
			r.logger.Warn("Error logging security event", log.Error(err))
		}

	}
	item, _ := httpcli.GetOutboundRequestLogItem(key)
	return &OutboundRequestResolver{req: item}, nil
}

func (r *outboundRequestConnectionResolver) Nodes(ctx context.Context) ([]*OutboundRequestResolver, error) {
	resolvers, err := r.compute(ctx)

	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		resolvers = resolvers[:*r.first]
	}

	return resolvers, nil
}

func (r *outboundRequestConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	resolvers, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(resolvers)), nil
}

func (r *outboundRequestConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	resolvers, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		return graphqlutil.NextPageCursor(string(resolvers[*r.first-1].ID())), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *outboundRequestConnectionResolver) compute(ctx context.Context) ([]*OutboundRequestResolver, error) {
	r.once.Do(func() {
		requests, err := httpcli.GetOutboundRequestLogItems(ctx, r.after)
		if err != nil {
			r.resolvers, r.err = nil, err
		}

		resolvers := make([]*OutboundRequestResolver, 0, len(requests))
		for _, item := range requests {
			resolvers = append(resolvers, &OutboundRequestResolver{req: item})
		}

		r.resolvers, r.err = resolvers, nil
	})
	return r.resolvers, r.err
}

func (r *OutboundRequestResolver) ID() graphql.ID {
	return relay.MarshalID("OutboundRequest", r.req.ID)
}

func (r *OutboundRequestResolver) StartedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.req.StartedAt}
}

func (r *OutboundRequestResolver) Method() string { return r.req.Method }

func (r *OutboundRequestResolver) URL() string { return r.req.URL }

func (r *OutboundRequestResolver) RequestHeaders() ([]*HttpHeaders, error) {
	return newHttpHeaders(r.req.RequestHeaders)
}

func (r *OutboundRequestResolver) RequestBody() string { return r.req.RequestBody }

func (r *OutboundRequestResolver) StatusCode() int32 { return r.req.StatusCode }

func (r *OutboundRequestResolver) ResponseHeaders() ([]*HttpHeaders, error) {
	return newHttpHeaders(r.req.ResponseHeaders)
}

func (r *OutboundRequestResolver) DurationMs() int32 { return int32(math.Round(r.req.Duration * 1000)) }

func (r *OutboundRequestResolver) ErrorMessage() string { return r.req.ErrorMessage }

func (r *OutboundRequestResolver) CreationStackFrame() string { return r.req.CreationStackFrame }

func (r *OutboundRequestResolver) CallStack() string { return r.req.CallStackFrame }

func newHttpHeaders(headers map[string][]string) ([]*HttpHeaders, error) {
	result := make([]*HttpHeaders, 0, len(headers))
	for key, values := range headers {
		result = append(result, &HttpHeaders{name: key, values: values})
	}

	return result, nil
}

func (h HttpHeaders) Name() string {
	return h.name
}

func (h HttpHeaders) Values() []string {
	return h.values
}

package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type outboundRequestsArgs struct {
	After *graphql.ID
}

type outboundRequestResolver struct {
	req *types.OutboundRequestLogItem
}

type HttpHeaders struct {
	name   string
	values []string
}

// accessTokenConnectionResolver resolves a list of access tokens.
//
// 🚨 SECURITY: When instantiating an outboundRequestConnectionResolver value, the caller MUST check
// permissions.
type outboundRequestConnectionResolver struct {
	after string
}

func (r *schemaResolver) OutboundRequests(ctx context.Context, args *outboundRequestsArgs) (*outboundRequestConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may list outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var after string
	if args.After != nil {
		err := relay.UnmarshalSpec(*args.After, &after)
		if err != nil {
			return nil, err
		}
	} else {
		after = ""
	}

	return &outboundRequestConnectionResolver{
		after: after,
	}, nil
}

func (r *schemaResolver) outboundRequestByID(ctx context.Context, id graphql.ID) (*outboundRequestResolver, error) {
	// 🚨 SECURITY: Only site admins may view outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var key string
	err := relay.UnmarshalSpec(id, key)
	if err != nil {
		return nil, err
	}
	item, _ := httpcli.GetOutboundRequestLogItem(key)
	return &outboundRequestResolver{req: item}, nil
}

func (r *outboundRequestConnectionResolver) Nodes(ctx context.Context) ([]*outboundRequestResolver, error) {
	requests, err := httpcli.GetAllOutboundRequestLogItemsAfter(ctx, r.after)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*outboundRequestResolver, 0, len(requests))
	for _, item := range requests {
		resolvers = append(resolvers, &outboundRequestResolver{req: item})
	}

	return resolvers, nil
}

func (r *outboundRequestConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	requests, err := httpcli.GetAllOutboundRequestLogItemsAfter(ctx, r.after)
	if err != nil {
		return 0, err
	}
	return int32(len(requests)), nil
}

func (r *outboundRequestConnectionResolver) PageInfo() (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (r *outboundRequestResolver) ID() graphql.ID {
	return relay.MarshalID("OutboundRequest", r.req.ID)
}

func (r *outboundRequestResolver) StartedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.req.StartedAt}
}

func (r *outboundRequestResolver) Method() string { return r.req.Method }

func (r *outboundRequestResolver) URL() string { return r.req.URL }

func (r *outboundRequestResolver) RequestHeaders() ([]*HttpHeaders, error) {
	return newHttpHeaders(r.req.RequestHeaders)
}

func (r *outboundRequestResolver) RequestBody() string { return r.req.RequestBody }

func (r *outboundRequestResolver) StatusCode() int32 { return r.req.StatusCode }

func (r *outboundRequestResolver) ResponseHeaders() ([]*HttpHeaders, error) {
	return newHttpHeaders(r.req.ResponseHeaders)
}

func (r *outboundRequestResolver) Duration() float64 { return r.req.Duration }

func (r *outboundRequestResolver) ErrorMessage() string { return r.req.ErrorMessage }

func (r *outboundRequestResolver) CreationStackFrame() string { return r.req.CreationStackFrame }

func (r *outboundRequestResolver) CallStackFrame() string { return r.req.CallStackFrame }

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

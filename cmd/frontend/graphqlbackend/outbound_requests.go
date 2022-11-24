package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type outboundRequestsArgs struct {
	After *string
}

type OutboundRequestResolver struct {
	req *types.OutboundRequestLogItem
}

type HttpHeaders struct {
	name   string
	values []string
}

func (r *schemaResolver) OutboundRequests(ctx context.Context, args *outboundRequestsArgs) ([]*OutboundRequestResolver, error) {
	// 🚨 SECURITY: Only site admins may list outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	result, err := httpcli.GetAllOutboundRequestLogItemsAfter(args.After, conf.Get().OutboundRequestLogLimit)
	if err != nil {
		return nil, err
	}

	var resolvers []*OutboundRequestResolver
	for _, item := range result {
		resolvers = append(resolvers, &OutboundRequestResolver{req: item})
	}

	return resolvers, nil
}

func (r *OutboundRequestResolver) Key() string {
	return r.req.Key
}

func (r *OutboundRequestResolver) StartedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.req.StartedAt}
}

func (r *OutboundRequestResolver) Method() string {
	return r.req.Method
}

func (r *OutboundRequestResolver) URL() string {
	return r.req.URL
}

func (r *OutboundRequestResolver) RequestHeaders() ([]*HttpHeaders, error) {
	return newHttpHeaders(r.req.RequestHeaders)
}

func (r *OutboundRequestResolver) RequestBody() string {
	return r.req.RequestBody
}

func (r *OutboundRequestResolver) StatusCode() int32 {
	return r.req.StatusCode
}

func (r *OutboundRequestResolver) ResponseHeaders() ([]*HttpHeaders, error) {
	return newHttpHeaders(r.req.ResponseHeaders)
}

func (r *OutboundRequestResolver) Duration() float64 {
	return r.req.Duration
}

func (r *OutboundRequestResolver) ErrorMessage() string {
	return r.req.ErrorMessage
}

func (r *OutboundRequestResolver) CreationStackFrame() string {
	return r.req.CreationStackFrame
}

func (r *OutboundRequestResolver) CallStackFrame() string {
	return r.req.CallStackFrame
}

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

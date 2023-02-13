package graphql

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rootResolver struct {
	sentinelSvc *sentinel.Service
	operations  *operations
}

func NewRootResolver(observationCtx *observation.Context, sentinelSvc *sentinel.Service) resolverstubs.SentinelServiceResolver {
	return &rootResolver{
		sentinelSvc: sentinelSvc,
		operations:  newOperations(observationCtx),
	}
}

func (r *rootResolver) Foo(ctx context.Context) (err error) {
	ctx, traceErrs, endObservation := r.operations.foo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	// ToDO
	_ = ctx
	_ = traceErrs
	return nil
}

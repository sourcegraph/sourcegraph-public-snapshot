package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver interface {
	PolicyResolverFactory(ctx context.Context) (_ PolicyResolver, err error)
}

type resolver struct {
	svc        Service
	operations *operations
}

func New(svc Service, observationContext *observation.Context) Resolver {
	return &resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

const slowQueryResolverRequestThreshold = time.Second

func (r *resolver) PolicyResolverFactory(ctx context.Context) (_ PolicyResolver, err error) {
	_, _, endObservation := observeResolver(ctx, &err, r.operations.getPolicyResolverFactory, slowQueryResolverRequestThreshold, observation.Args{})
	defer endObservation()

	return NewPolicyResolver(r.svc, r.operations), nil
}

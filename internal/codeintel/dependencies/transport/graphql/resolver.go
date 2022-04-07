package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *dependencies.Service
	operations *operations
}

func newResolver(svc *dependencies.Service, observationContext *observation.Context) *Resolver {
	return &Resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *Resolver) Dependencies(ctx context.Context) (err error) {
	ctx, endObservation := r.operations.dependencies.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	_, err = r.svc.Dependencies(ctx, nil)
	return err
}

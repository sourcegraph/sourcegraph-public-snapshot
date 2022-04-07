package graphql

import (
	"context"

	autoindexing "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *autoindexing.Service
	operations *operations
}

func newResolver(svc *autoindexing.Service, observationContext *observation.Context) *Resolver {
	return &Resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *Resolver) Todo(ctx context.Context) (err error) {
	ctx, endObservation := r.operations.todo.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.svc.Todo(ctx)
}

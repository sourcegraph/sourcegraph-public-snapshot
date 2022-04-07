package graphql

import (
	"context"

	documents "github.com/sourcegraph/sourcegraph/internal/codeintel/documents"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *documents.Service
	operations *operations
}

func newResolver(svc *documents.Service, observationContext *observation.Context) *Resolver {
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

package graphql

import (
	"context"

	documents "github.com/sourcegraph/sourcegraph/internal/codeintel/documents"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (r *Resolver) Document(ctx context.Context, args struct{}) (_ any, err error) {
	ctx, _, endObservation := r.operations.document.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33373
	_, _ = ctx, args
	return nil, errors.New("unimplemented: Document")
}

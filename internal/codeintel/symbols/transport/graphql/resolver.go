package graphql

import (
	"context"
	"errors"

	symbols "github.com/sourcegraph/sourcegraph/internal/codeintel/symbols"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Resolver struct {
	svc        *symbols.Service
	operations *operations
}

func newResolver(svc *symbols.Service, observationContext *observation.Context) *Resolver {
	return &Resolver{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (r *Resolver) Symbol(ctx context.Context, args struct{}) (_ any, err error) {
	ctx, _, endObservation := r.operations.symbol.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33374
	_, _ = ctx, args
	return nil, errors.New("unimplemented: Symbol")
}

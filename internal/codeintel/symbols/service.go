package symbols

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	symbolsStore store.Store
	operations   *operations
}

func newService(symbolsStore store.Store, observationContext *observation.Context) *Service {
	return &Service{
		symbolsStore: symbolsStore,
		operations:   newOperations(observationContext),
	}
}

type Symbol = shared.Symbol

type SymbolOpts struct{}

func (s *Service) Symbol(ctx context.Context, opts SymbolOpts) (symbols []Symbol, err error) {
	ctx, _, endObservation := s.operations.symbol.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33374
	_, _ = s.symbolsStore.List(ctx, store.ListOpts{})
	return nil, errors.Newf("unimplemented: symbols.Symbol")
}

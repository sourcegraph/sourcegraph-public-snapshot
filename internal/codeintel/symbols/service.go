package symbols

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	symbolsStore Store
	operations   *operations
}

func newService(symbolsStore Store, observationContext *observation.Context) *Service {
	return &Service{
		symbolsStore: symbolsStore,
		operations:   newOperations(observationContext),
	}
}

type Symbol struct {
	// TODO
}

type SymbolOpts struct {
	// TODO
}

func (s *Service) Symbol(ctx context.Context, opts SymbolOpts) (symbols []Symbol, err error) {
	ctx, endObservation := s.operations.symbol.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	return nil, nil
}

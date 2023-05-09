package context

import (
	"context"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	store      store.Store
	operations *operations
}

func newService(
	observationCtx *observation.Context,
	store store.Store,
) *Service {
	return &Service{
		store:      store,
		operations: newOperations(observationCtx),
	}
}

func (s *Service) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
	return "", nil
}

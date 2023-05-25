package context

import (
	"context"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

	scipstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/scipstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	scipstore  scipstore.ScipStore
	operations *operations
}

func newService(
	observationCtx *observation.Context,
	scipstore scipstore.ScipStore,
) *Service {
	return &Service{
		scipstore:  scipstore,
		operations: newOperations(observationCtx),
	}
}

func (s *Service) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
	return "", nil
}

package context

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/context/internal/store"
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

func (s *Service) SplitIntoEmbeddableChunks(ctx context.Context, text string, fileName string, splitOptions SplitOptions) ([]EmbeddableChunk, error) {
	return SplitIntoEmbeddableChunks(text, fileName, splitOptions), nil
}

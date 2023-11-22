package embed

import (
	"context"

	codeintelContext "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
)

type ContextService interface {
	SplitIntoEmbeddableChunks(ctx context.Context, text string, fileName string, splitOptions codeintelContext.SplitOptions) ([]codeintelContext.EmbeddableChunk, error)
}

package embed

import (
	"context"

	codeintelContext "github.com/sourcegraph/sourcegraph/internal/codeintel/codycontext"
)

type ContextService interface {
	SplitIntoEmbeddableChunks(ctx context.Context, text string, fileName string, splitOptions codeintelContext.SplitOptions) ([]codeintelContext.EmbeddableChunk, error)
}

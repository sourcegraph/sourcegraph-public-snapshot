pbckbge embed

import (
	"context"

	codeintelContext "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context"
)

type ContextService interfbce {
	SplitIntoEmbeddbbleChunks(ctx context.Context, text string, fileNbme string, splitOptions codeintelContext.SplitOptions) ([]codeintelContext.EmbeddbbleChunk, error)
}

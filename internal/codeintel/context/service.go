pbckbge context

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Service struct {
	store      store.Store
	operbtions *operbtions
}

func newService(
	observbtionCtx *observbtion.Context,
	store store.Store,
) *Service {
	return &Service{
		store:      store,
		operbtions: newOperbtions(observbtionCtx),
	}
}

func (s *Service) SplitIntoEmbeddbbleChunks(ctx context.Context, text string, fileNbme string, splitOptions SplitOptions) ([]EmbeddbbleChunk, error) {
	return SplitIntoEmbeddbbleChunks(text, fileNbme, splitOptions), nil
}

pbckbge embeddings

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/embeddings/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func Init(
	_ context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	repoEmbeddingsStore := repo.NewRepoEmbeddingJobsStore(db)
	gitserverClient := gitserver.NewClient()
	embeddingsClient := embeddings.NewDefbultClient()
	enterpriseServices.EmbeddingsResolver = resolvers.NewResolver(
		db,
		observbtionCtx.Logger,
		gitserverClient,
		embeddingsClient,
		repoEmbeddingsStore,
	)

	return nil
}

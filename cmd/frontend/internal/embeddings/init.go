package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/embeddings/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	repoEmbeddingsStore := repo.NewRepoEmbeddingJobsStore(db)
	gitserverClient := gitserver.NewClient("graphql.embeddings")
	embeddingsClient := embeddings.NewDefaultClient()
	enterpriseServices.EmbeddingsResolver = resolvers.NewResolver(
		db,
		observationCtx.Logger,
		gitserverClient,
		embeddingsClient,
		repoEmbeddingsStore,
	)

	return nil
}

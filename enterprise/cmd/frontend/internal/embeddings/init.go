package embeddings

import (
	"context"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/embeddings/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	repoEmbeddingsStore := repo.NewRepoEmbeddingJobsStore(db)
	contextDetectionEmbeddingsStore := contextdetection.NewContextDetectionEmbeddingJobsStore(db)
	gitserverClient := gitserver.NewClient()
	embeddingsClient := embeddings.NewClient()
	// TODO (stefan): Do we need the StartupTimeout?
	weaviateClient, err := weaviate.NewClient(weaviate.Config{Host: "localhost:8181", Scheme: "http", StartupTimeout: time.Second * 10})
	if err != nil {
		panic(err)
	}

	enterpriseServices.EmbeddingsResolver = resolvers.NewResolver(db, gitserverClient, embeddingsClient, repoEmbeddingsStore, contextDetectionEmbeddingsStore, weaviateClient)
	return nil
}

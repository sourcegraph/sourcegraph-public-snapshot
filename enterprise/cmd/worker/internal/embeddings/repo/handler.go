package repo

import (
	"context"

	"github.com/sourcegraph/log"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	repoembeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
}

var _ workerutil.Handler[*repoembeddingsbg.RepoEmbeddingJob] = &handler{}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *repoembeddingsbg.RepoEmbeddingJob) error {
	if !conf.EmbeddingsEnabled() {
		return errors.New("embeddings are not configured or disabled")
	}

	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	embeddingsClient := embed.NewEmbeddingsClient()

	repoEmbeddingIndex, stats, err := embed.EmbedRepo(
		ctx,
		repo.Name,
		record.Revision,
		embeddingsClient,
		h.gitserverClient,
	)
	if err != nil {
		return err
	}

	logger.Info(
		"finished generating repo embeddings",
		log.String("repoName", string(repo.Name)),
		log.String("revision", string(record.Revision)),
		log.Object("stats", stats.ToFields()...),
	)

	return embeddings.UploadRepoEmbeddingIndex(ctx, h.uploadStore, string(embeddings.GetRepoEmbeddingIndexName(repo.Name)), repoEmbeddingIndex)
}

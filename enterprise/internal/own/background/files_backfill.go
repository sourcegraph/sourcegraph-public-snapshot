package background

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func handleFilesBackfill(ctx context.Context, lgr logger.Logger, repoId api.RepoID, db database.DB) error {
	// 🚨 SECURITY: we use the internal actor because the background indexer is not associated with any user, and needs
	// to see all repos and files
	internalCtx := actor.WithInternalActor(ctx)
	lgr.Info("backfilling files for repository")
	indexer := newFilesBackfillIndexer(gitserver.NewClient(), db, lgr)
	return indexer.indexRepo(internalCtx, repoId)
}

type filesBackfillIndexer struct {
	client gitserver.Client
	db     database.DB
	logger logger.Logger
}

func newFilesBackfillIndexer(client gitserver.Client, db database.DB, lgr logger.Logger) *filesBackfillIndexer {
	return &filesBackfillIndexer{client: client, db: db, logger: lgr}
}

var filesCounter = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "own_files_backfill_files_indexed_total",
})

func (r *filesBackfillIndexer) indexRepo(ctx context.Context, repoId api.RepoID) error {
	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrap(err, "repoStore.Get")
	}
	r.logger.Info("LsFines", logger.String("repo_name", string(repo.Name)))
	files, err := r.client.LsFiles(ctx, nil, repo.Name, "HEAD")
	if err != nil {
		r.logger.Error("ls-files failed", logger.String("msg", err.Error()))
		return errors.Wrap(err, "LsFiles")
	}
	newlyInserted, err := r.db.RepoPaths().EnsureExist(ctx, repo.ID, files)
	if err != nil {
		r.logger.Error("inserting backfill files failed", logger.String("msg", err.Error()))
		return errors.Wrap(err, "EnsureExist")
	}
	r.logger.Info("files", logger.Int("total", len(files)), logger.Int("diff", newlyInserted), logger.String("repo_name", string(repo.Name)))
	filesCounter.Add(float64(len(files)))
	return nil
}

package background

import (
	"context"
	"fmt"

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
	// commitLog, err := r.client.CommitLog(ctx, repo.Name, time.Now().AddDate(0, 0, -90))
	// if err != nil {
	// 	return errors.Wrap(err, "CommitLog")
	// }
	files, err := r.client.LsFiles(ctx, nil, repo.Name, "HEAD")
	if err != nil {
		return err
	}
	for _, f := range files {
		fmt.Println(f)
	}
	r.logger.Info("commits inserted", logger.Int("count", len(files)), logger.Int("repo_id", int(repoId)))
	commitCounter.Add(float64(len(files)))
	return nil
}

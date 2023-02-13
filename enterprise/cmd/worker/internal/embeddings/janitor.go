package embeddings

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	embeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type repoEmbeddingJanitorJob struct{}

func NewRepoEmbeddingJanitorJob() job.Job {
	return &repoEmbeddingJanitorJob{}
}

func (j *repoEmbeddingJanitorJob) Description() string {
	return ""
}

func (j *repoEmbeddingJanitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *repoEmbeddingJanitorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	// TODO: Check if embeddings are enabled
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	store := embeddingsbg.NewRepoEmbeddingJobWorkerStore(observationCtx, db.Handle())
	return []goroutine.BackgroundRoutine{newRepoEmbeddingJobResetter(observationCtx, store)}, nil
}

func newRepoEmbeddingJobResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*embeddingsbg.RepoEmbeddingJob]) *dbworker.Resetter[*embeddingsbg.RepoEmbeddingJob] {
	return dbworker.NewResetter(observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "embedding_job_worker_resetter",
		Interval: time.Minute * 60, // Check for orphaned jobs every 60 minutes
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "embedding_job_worker"),
	})
}

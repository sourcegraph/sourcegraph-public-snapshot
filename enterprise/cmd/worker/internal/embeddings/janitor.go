package embeddings

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type embeddingJanitorJob struct{}

func NewEmbeddingJanitorJob() job.Job {
	return &embeddingJanitorJob{}
}

func (j *embeddingJanitorJob) Description() string {
	return ""
}

func (j *embeddingJanitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *embeddingJanitorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	// TODO: Check if embeddings are enabled
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	store := newEmbeddingJobWorkerStore(observationCtx, db.Handle())
	return []goroutine.BackgroundRoutine{newEmbeddingJobResetter(observationCtx, store)}, nil
}

func newEmbeddingJobResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*EmbeddingJob]) *dbworker.Resetter[*EmbeddingJob] {
	return dbworker.NewResetter(observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "embedding_job_worker_resetter",
		Interval: time.Minute * 60, // Check for orphaned jobs every 60 minutes
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "embedding_job_worker"),
	})
}

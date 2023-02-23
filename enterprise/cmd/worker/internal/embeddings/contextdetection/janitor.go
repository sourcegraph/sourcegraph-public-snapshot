package contextdetection

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	contextdetectionbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type contextDetectionEmbeddingJanitorJob struct{}

func NewContextDetectionEmbeddingJanitorJob() job.Job {
	return &contextDetectionEmbeddingJanitorJob{}
}

func (j *contextDetectionEmbeddingJanitorJob) Description() string {
	return ""
}

func (j *contextDetectionEmbeddingJanitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *contextDetectionEmbeddingJanitorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	store := contextdetectionbg.NewContextDetectionEmbeddingJobWorkerStore(observationCtx, db.Handle())
	return []goroutine.BackgroundRoutine{newContextDetectionEmbeddingJobResetter(observationCtx, store)}, nil
}

func newContextDetectionEmbeddingJobResetter(observationCtx *observation.Context, workerStore dbworkerstore.Store[*contextdetectionbg.ContextDetectionEmbeddingJob]) *dbworker.Resetter[*contextdetectionbg.ContextDetectionEmbeddingJob] {
	return dbworker.NewResetter(observationCtx.Logger, workerStore, dbworker.ResetterOptions{
		Name:     "context_detection_embedding_job_worker_resetter",
		Interval: time.Minute, // Check for orphaned jobs every minute
		Metrics:  dbworker.NewResetterMetrics(observationCtx, "context_detection_embedding_job_worker"),
	})
}

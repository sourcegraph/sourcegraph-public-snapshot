package contextdetection

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	contextdetectionbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type contextDetectionEmbeddingJob struct{}

func NewContextDetectionEmbeddingJob() job.Job {
	return &contextDetectionEmbeddingJob{}
}

func (s *contextDetectionEmbeddingJob) Description() string {
	return ""
}

func (s *contextDetectionEmbeddingJob) Config() []env.Config {
	return []env.Config{embeddings.EmbeddingsUploadStoreConfigInst}
}

func (s *contextDetectionEmbeddingJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	uploadStore, err := embeddings.NewEmbeddingsUploadStore(context.Background(), observationCtx, embeddings.EmbeddingsUploadStoreConfigInst)
	if err != nil {
		return nil, err
	}

	workCtx := actor.WithInternalActor(context.Background())
	return []goroutine.BackgroundRoutine{
		newContextDetectionEmbeddingJobWorker(
			workCtx,
			observationCtx,
			contextdetectionbg.NewContextDetectionEmbeddingJobWorkerStore(observationCtx, db.Handle()),
			edb.NewEnterpriseDB(db),
			uploadStore,
			gitserver.NewClient(),
		),
	}, nil
}

func newContextDetectionEmbeddingJobWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*contextdetectionbg.ContextDetectionEmbeddingJob],
	db edb.EnterpriseDB,
	uploadStore uploadstore.Store,
	gitserverClient gitserver.Client,
) *workerutil.Worker[*contextdetectionbg.ContextDetectionEmbeddingJob] {
	handler := &handler{db, uploadStore, gitserverClient}
	return dbworker.NewWorker[*contextdetectionbg.ContextDetectionEmbeddingJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "context_detection_embedding_job_worker",
		Interval:          time.Minute, // Poll for a job once per minute
		NumHandlers:       1,           // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "context_detection_embedding_job_worker"),
	})
}

package file

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	fileembeddingsbg "github.com/sourcegraph/sourcegraph/internal/embeddings/background/file"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type fileEmbeddingJob struct{}

func NewFileEmbeddingJob() job.Job {
	return &fileEmbeddingJob{}
}

func (s *fileEmbeddingJob) Description() string {
	return ""
}

func (s *fileEmbeddingJob) Config() []env.Config {
	return []env.Config{embeddings.EmbeddingsUploadStoreConfigInst}
}

func (s *fileEmbeddingJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	uploadStore, err := embeddings.NewEmbeddingsUploadStore(context.Background(), observationCtx, embeddings.EmbeddingsUploadStoreConfigInst)
	if err != nil {
		return nil, err
	}

	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	workCtx := actor.WithInternalActor(context.Background())
	return []goroutine.BackgroundRoutine{
		newFileEmbeddingJobWorker(
			workCtx,
			observationCtx,
			fileembeddingsbg.NewFileEmbeddingJobWorkerStore(observationCtx, db.Handle()),
			db,
			uploadStore,
			gitserver.NewClient(),
			services.ContextService,
			fileembeddingsbg.NewFileEmbeddingJobsStore(db),
		),
	}, nil
}

func newFileEmbeddingJobWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*fileembeddingsbg.FileEmbeddingJob],
	db database.DB,
	uploadStore uploadstore.Store,
	gitserverClient gitserver.Client,
	contextService embed.ContextService,
	fileEmbeddingJobsStore fileembeddingsbg.FileEmbeddingJobsStore,
) *workerutil.Worker[*fileembeddingsbg.FileEmbeddingJob] {
	handler := &handler{
		db:                     db,
		uploadStore:            uploadStore,
		gitserverClient:        gitserverClient,
		contextService:         contextService,
		fileEmbeddingJobsStore: fileEmbeddingJobsStore,
	}
	return dbworker.NewWorker[*fileembeddingsbg.FileEmbeddingJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "file_embedding_job_worker",
		Interval:          10 * time.Second, // Poll for a job once every 10 seconds
		NumHandlers:       1,                // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "file_embedding_job_worker"),
	})
}

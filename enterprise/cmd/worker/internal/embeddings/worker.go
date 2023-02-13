package embeddings

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type embeddingJob struct{}

func NewEmbeddingJob() job.Job {
	return &embeddingJob{}
}

func (s *embeddingJob) Description() string {
	return ""
}

func (s *embeddingJob) Config() []env.Config {
	return []env.Config{embeddings.EmbeddingsUploadStoreConfigInst}
}

func (s *embeddingJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	// TODO: Check if embeddings are enabled
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	enterpriseDB := edb.NewEnterpriseDB(db)
	store := newEmbeddingJobWorkerStore(observationCtx, db.Handle())
	workCtx := actor.WithInternalActor(context.Background())
	uploadStore, err := embeddings.NewEmbeddingsUploadStore(workCtx, observationCtx, embeddings.EmbeddingsUploadStoreConfigInst)
	if err != nil {
		return nil, err
	}
	gitserverClient := gitserver.NewClient()
	return []goroutine.BackgroundRoutine{newEmbeddingJobWorker(workCtx, observationCtx, store, enterpriseDB, uploadStore, gitserverClient)}, nil
}

func newEmbeddingJobWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*EmbeddingJob],
	db edb.EnterpriseDB,
	uploadStore uploadstore.Store,
	gitserverClient gitserver.Client,
) *workerutil.Worker[*EmbeddingJob] {
	mu := sync.Mutex{}
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// c := conf.Get()
		// Get the list of repos
		// Check if repos are valid
		// Get latest revision for each repo
		// Insert them into the jobs table
	})

	handler := &handler{db, uploadStore, gitserverClient}
	return dbworker.NewWorker[*EmbeddingJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "embedding_job_worker",
		Interval:          time.Second, // Poll for a job once per second
		NumHandlers:       1,           // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "embedding_job_worker"),
	})
}

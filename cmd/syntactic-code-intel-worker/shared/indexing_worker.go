package shared

import (
	"context"
	"github.com/sourcegraph/log"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewIndexingWorker(ctx context.Context, observationCtx *observation.Context, workerStore dbworkerstore.Store[*SyntacticIndexingJob], config IndexingWorkerConfig) *workerutil.Worker[*SyntacticIndexingJob] {

	name := "syntactic_code_intel_indexing_worker"

	return dbworker.NewWorker[*SyntacticIndexingJob](ctx, workerStore, &indexingHandler{}, workerutil.WorkerOptions{
		Name:                 name,
		Interval:             config.PollInterval,
		HeartbeatInterval:    10 * time.Second,
		Metrics:              workerutil.NewMetrics(observationCtx, name),
		NumHandlers:          config.Concurrency,
		MaximumRuntimePerJob: config.MaximumRuntimePerJob,
	})

}

type indexingHandler struct{}

var _ workerutil.Handler[*SyntacticIndexingJob] = &indexingHandler{}

func (i indexingHandler) Handle(ctx context.Context, logger log.Logger, record *SyntacticIndexingJob) error {
	logger.Info("Stub indexing worker handling record",
		log.Int("id", record.ID),
		log.String("repository name", record.RepositoryName),
		log.String("commit", record.Commit))
	return nil
}

package shared

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewIndexingWorker(ctx context.Context,
	observationCtx *observation.Context,
	jobStore jobstore.SyntacticIndexingJobStore,
	config IndexingWorkerConfig) *workerutil.Worker[*jobstore.SyntacticIndexingJob] {

	name := "syntactic_code_intel_indexing_worker"

	return dbworker.NewWorker(ctx, jobStore.DBWorkerStore(), &indexingHandler{}, workerutil.WorkerOptions{
		Name:                 name,
		Interval:             config.PollInterval,
		HeartbeatInterval:    10 * time.Second,
		Metrics:              workerutil.NewMetrics(observationCtx, name),
		NumHandlers:          config.Concurrency,
		MaximumRuntimePerJob: config.MaximumRuntimePerJob,
	})

}

type indexingHandler struct{}

var _ workerutil.Handler[*jobstore.SyntacticIndexingJob] = &indexingHandler{}

func (i indexingHandler) Handle(ctx context.Context, logger log.Logger, record *jobstore.SyntacticIndexingJob) error {
	logger.Info("Stub indexing worker handling record",
		log.Int("id", record.ID),
		log.String("repository name", record.RepositoryName),
		log.String("commit", string(record.Commit)))
	return nil
}

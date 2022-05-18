package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/indexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type indexingJob struct{}

func NewIndexingJob() job.Job {
	return &indexingJob{}
}

func (j *indexingJob) Description() string {
	return ""
}

func (j *indexingJob) Config() []env.Config {
	return []env.Config{indexingConfigInst}
}

func (j *indexingJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "indexing job routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	dependencySyncStore, err := codeintel.InitDependencySyncingStore()
	if err != nil {
		return nil, err
	}

	dependencyIndexingStore, err := codeintel.InitDependencyIndexingStore()
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	dbworker.InitPrometheusMetric(observationContext, dependencySyncStore, "codeintel", "dependency_index", nil)

	repoUpdaterClient := codeintel.InitRepoUpdaterClient()
	extSvcStore := database.NewDB(db).ExternalServices()
	dbStoreShim := &indexing.DBStoreShim{Store: dbStore}
	enqueuerDBStoreShim := &enqueuer.DBStoreShim{Store: dbStore}
	policyMatcher := policies.NewMatcher(gitserverClient, policies.IndexingExtractor, false, true)
	syncMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor")
	queueingMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing")
	indexEnqueuer := enqueuer.NewIndexEnqueuer(enqueuerDBStoreShim, gitserverClient, repoUpdaterClient, indexingConfigInst.AutoIndexEnqueuerConfig, observationContext)

	routines := []goroutine.BackgroundRoutine{
		indexing.NewIndexScheduler(dbStoreShim, policyMatcher, indexEnqueuer, indexingConfigInst.RepositoryProcessDelay, indexingConfigInst.RepositoryBatchSize, indexingConfigInst.PolicyBatchSize, indexingConfigInst.AutoIndexingTaskInterval, observationContext),
		indexing.NewDependencySyncScheduler(dbStoreShim, dependencySyncStore, extSvcStore, syncMetrics),
		indexing.NewDependencyIndexingScheduler(dbStoreShim, dependencyIndexingStore, extSvcStore, repoUpdaterClient, gitserverClient, indexEnqueuer, indexingConfigInst.DependencyIndexerSchedulerPollInterval, indexingConfigInst.DependencyIndexerSchedulerConcurrency, queueingMetrics),
	}

	return routines, nil
}

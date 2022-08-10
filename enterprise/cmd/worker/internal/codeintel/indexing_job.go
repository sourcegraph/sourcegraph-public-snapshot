package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/indexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
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
	databaseDB := database.NewDB(logger, db)
	extSvcStore := databaseDB.ExternalServices()
	gitserverRepoStore := databaseDB.GitserverRepos()
	dbStoreShim := &indexing.DBStoreShim{Store: dbStore}
	enqueuerDBStoreShim := &autoindexing.DBStoreShim{Store: dbStore}
	syncMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor")
	queueingMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing")
	indexEnqueuer := autoindexing.GetService(databaseDB, enqueuerDBStoreShim, gitserverClient, repoUpdaterClient)

	routines := []goroutine.BackgroundRoutine{
		indexing.NewDependencySyncScheduler(dbStoreShim, dependencySyncStore, extSvcStore, syncMetrics, observationContext),
		indexing.NewDependencyIndexingScheduler(dbStoreShim, dependencyIndexingStore, extSvcStore, gitserverRepoStore, repoUpdaterClient, indexEnqueuer, indexingConfigInst.DependencyIndexerSchedulerPollInterval, indexingConfigInst.DependencyIndexerSchedulerConcurrency, queueingMetrics),
	}

	return routines, nil
}

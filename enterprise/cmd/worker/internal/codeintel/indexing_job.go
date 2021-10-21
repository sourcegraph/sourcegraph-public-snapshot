package codeintel

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/indexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type indexingJob struct{}

func NewIndexingJob() shared.Job {
	return &indexingJob{}
}

func (j *indexingJob) Config() []env.Config {
	return []env.Config{indexingConfigInst}
}

func (j *indexingJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := shared.InitDatabase()
	if err != nil {
		return nil, err
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	gitserverClient, err := InitGitserverClient()
	if err != nil {
		return nil, err
	}

	dependencySyncStore, err := InitDependencySyncingStore()
	if err != nil {
		return nil, err
	}

	dependencyIndexingStore, err := InitDependencyIndexingStore()
	if err != nil {
		return nil, err
	}

	extSvcStore := database.ExternalServices(db)
	dbStoreShim := &indexing.DBStoreShim{Store: dbStore}
	enqueuerDBStoreShim := &enqueuer.DBStoreShim{Store: dbStore}
	policyMatcher := policies.NewMatcher(gitserverClient, policies.IndexingExtractor, false, true)
	indexEnqueuer := enqueuer.NewIndexEnqueuer(enqueuerDBStoreShim, gitserverClient, repoupdater.DefaultClient, indexingConfigInst.AutoIndexEnqueuerConfig, observationContext)
	syncMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor", nil)
	queueingMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing", nil)

	prometheus.DefaultRegisterer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_dependency_index_total",
		Help: "Total number of jobs in the queued state.",
	}, func() float64 {
		count, err := dependencySyncStore.QueuedCount(context.Background(), false, nil)
		if err != nil {
			log15.Error("Failed to get queued job count", "error", err)
		}

		return float64(count)
	}))

	routines := []goroutine.BackgroundRoutine{
		indexing.NewIndexScheduler(dbStoreShim, policyMatcher, indexEnqueuer, indexingConfigInst.RepositoryProcessDelay, indexingConfigInst.RepositoryBatchSize, indexingConfigInst.AutoIndexingTaskInterval, observationContext),
		indexing.NewDependencySyncScheduler(dbStoreShim, dependencySyncStore, extSvcStore, syncMetrics),
		indexing.NewDependencyIndexingScheduler(dbStoreShim, dependencyIndexingStore, extSvcStore, repoupdater.DefaultClient, gitserverClient, indexEnqueuer, indexingConfigInst.DependencyIndexerSchedulerPollInterval, indexingConfigInst.DependencyIndexerSchedulerConcurrency, queueingMetrics),
	}

	return routines, nil
}

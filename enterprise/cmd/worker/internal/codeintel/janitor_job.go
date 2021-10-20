package codeintel

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type janitorJob struct{}

func NewJanitorJob() shared.Job {
	return &janitorJob{}
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	lsifStore, err := InitLSIFStore()
	if err != nil {
		return nil, err
	}

	dependencyIndexingStore, err := InitDependencySyncingStore()
	if err != nil {
		return nil, err
	}

	gitserverClient, err := InitGitserverClient()
	if err != nil {
		return nil, err
	}

	dbStoreShim := &janitor.DBStoreShim{Store: dbStore}
	lsifStoreShim := &janitor.LSIFStoreShim{Store: lsifStore}
	policyMatcher := policies.NewMatcher(gitserverClient, policies.RetentionExtractor, true, false)
	uploadWorkerStore := dbstore.WorkerutilUploadStore(dbStoreShim, observationContext)
	indexWorkerStore := dbstore.WorkerutilIndexStore(dbStoreShim, observationContext)
	metrics := janitor.NewMetrics(observationContext)

	executorMetricsReporter, err := executorqueue.NewMetricReporter(observationContext, "codeintel", indexWorkerStore, janitorConfigInst.MetricsConfig)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{
		// Reconciliation
		janitor.NewDeletedRepositoryJanitor(dbStoreShim, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewUnknownCommitJanitor(dbStoreShim, janitorConfigInst.CommitResolverMinimumTimeSinceLastCheck, janitorConfigInst.CommitResolverBatchSize, janitorConfigInst.CommitResolverTaskInterval, metrics),

		// Expiration
		janitor.NewAbandonedUploadJanitor(dbStoreShim, janitorConfigInst.UploadTimeout, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewUploadExpirer(dbStoreShim, policyMatcher, janitorConfigInst.RepositoryProcessDelay, janitorConfigInst.RepositoryBatchSize, janitorConfigInst.UploadProcessDelay, janitorConfigInst.UploadBatchSize, janitorConfigInst.CommitBatchSize, janitorConfigInst.BranchesCacheMaxKeys, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewExpiredUploadDeleter(dbStoreShim, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewHardDeleter(dbStoreShim, lsifStoreShim, janitorConfigInst.CleanupTaskInterval, metrics),

		// Current indexes
		janitor.NewDocumentationSearchCurrentJanitor(lsifStoreShim, janitorConfigInst.DocumentationSearchCurrentMinimumTimeSinceLastCheck, janitorConfigInst.DocumentationSearchCurrentBatchSize, janitorConfigInst.CleanupTaskInterval, metrics),

		// Resetters
		janitor.NewUploadResetter(uploadWorkerStore, janitorConfigInst.CleanupTaskInterval, metrics, observationContext),
		janitor.NewIndexResetter(indexWorkerStore, janitorConfigInst.CleanupTaskInterval, metrics, observationContext),
		janitor.NewDependencyIndexResetter(dependencyIndexingStore, janitorConfigInst.CleanupTaskInterval, metrics, observationContext),

		executorMetricsReporter,
	}

	return routines, nil
}

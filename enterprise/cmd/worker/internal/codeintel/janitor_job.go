package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/executorqueue"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type janitorJob struct{}

func NewJanitorJob() job.Job {
	return &janitorJob{}
}

func (j *janitorJob) Description() string {
	return ""
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{janitorConfigInst}
}

func (j *janitorJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "janitor job routines"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return nil, err
	}

	dependencyIndexingStore, err := codeintel.InitDependencySyncingStore()
	if err != nil {
		return nil, err
	}

	gitserverClient, err := codeintel.InitGitserverClient()
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
		// Reconciliation and denormalization
		janitor.NewDeletedRepositoryJanitor(dbStoreShim, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewUnknownCommitJanitor(dbStoreShim, janitorConfigInst.CommitResolverMinimumTimeSinceLastCheck, janitorConfigInst.CommitResolverBatchSize, janitorConfigInst.CommitResolverMaximumCommitLag, janitorConfigInst.CommitResolverTaskInterval, metrics),
		janitor.NewRepositoryPatternMatcher(dbStoreShim, lsifStoreShim, janitorConfigInst.CleanupTaskInterval, janitorConfigInst.ConfigurationPolicyMembershipBatchSize, metrics),

		// Expiration
		janitor.NewAbandonedUploadJanitor(dbStoreShim, janitorConfigInst.UploadTimeout, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewUploadExpirer(dbStoreShim, policyMatcher, janitorConfigInst.RepositoryProcessDelay, janitorConfigInst.RepositoryBatchSize, janitorConfigInst.UploadProcessDelay, janitorConfigInst.UploadBatchSize, janitorConfigInst.PolicyBatchSize, janitorConfigInst.CommitBatchSize, janitorConfigInst.BranchesCacheMaxKeys, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewExpiredUploadDeleter(dbStoreShim, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewHardDeleter(dbStoreShim, lsifStoreShim, janitorConfigInst.CleanupTaskInterval, metrics),
		janitor.NewAuditLogJanitor(dbStoreShim, janitorConfigInst.AuditLogMaxAge, janitorConfigInst.CleanupTaskInterval, metrics),

		// Resetters
		janitor.NewUploadResetter(uploadWorkerStore, janitorConfigInst.CleanupTaskInterval, metrics, observationContext),
		janitor.NewIndexResetter(indexWorkerStore, janitorConfigInst.CleanupTaskInterval, metrics, observationContext),
		janitor.NewDependencyIndexResetter(dependencyIndexingStore, janitorConfigInst.CleanupTaskInterval, metrics, observationContext),

		executorMetricsReporter,
	}

	return routines, nil
}

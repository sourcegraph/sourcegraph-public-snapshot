package autoindexing

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	autoindexingstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var (
	IndexWorkerStoreOptions                 = background.IndexWorkerStoreOptions
	DependencySyncingJobWorkerStoreOptions  = background.DependencySyncingJobWorkerStoreOptions
	DependencyIndexingJobWorkerStoreOptions = background.DependencyIndexingJobWorkerStoreOptions
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	depsSvc DependenciesService,
	policiesSvc PoliciesService,
	gitserverClient gitserver.Client,
) *Service {
	store := autoindexingstore.New(scopedContext("store", observationCtx), db)
	repoUpdater := repoupdater.DefaultClient
	inferenceSvc := inference.NewService()

	svc := newService(scopedContext("service", observationCtx), store, inferenceSvc, repoUpdater, db.Repos(), gitserverClient)

	return svc
}

func scopedContext(component string, observationCtx *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "autoindexing", component, observationCtx)
}

func NewResetters(observationCtx *observation.Context, db database.DB) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationCtx)
	indexStore := dbworkerstore.New(observationCtx, db.Handle(), background.IndexWorkerStoreOptions)
	dependencyIndexingStore := dbworkerstore.New(observationCtx, db.Handle(), background.DependencyIndexingJobWorkerStoreOptions)

	return []goroutine.BackgroundRoutine{
		background.NewIndexResetter(observationCtx.Logger.Scoped("indexResetter", ""), ConfigCleanupInst.Interval, indexStore, metrics),
		background.NewDependencyIndexResetter(observationCtx.Logger.Scoped("dependencyIndexResetter", ""), ConfigCleanupInst.Interval, dependencyIndexingStore, metrics),
	}
}

func NewJanitorJobs(observationCtx *observation.Context, autoindexingSvc *Service, gitserverClient gitserver.Client) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewUnknownRepositoryJanitor(
			autoindexingSvc.store,
			ConfigCleanupInst.Interval,
			observationCtx,
		),

		background.NewUnknownCommitJanitor(
			autoindexingSvc.store,
			gitserverClient,
			ConfigCleanupInst.Interval,
			ConfigCleanupInst.CommitResolverBatchSize,
			ConfigCleanupInst.MinimumTimeSinceLastCheck,
			ConfigCleanupInst.CommitResolverMaximumCommitLag,
			observationCtx,
		),

		background.NewExpiredRecordJanitor(
			autoindexingSvc.store,
			ConfigCleanupInst.Interval,
			ConfigCleanupInst.FailedIndexBatchSize,
			ConfigCleanupInst.FailedIndexMaxAge,
			observationCtx,
		),
	}
}

func NewIndexSchedulers(
	observationCtx *observation.Context,
	uploadSvc UploadService,
	policiesSvc PoliciesService,
	policyMatcher PolicyMatcher,
	autoindexingSvc *Service,
	repoStore database.RepoStore,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewScheduler(
			observationCtx,
			uploadSvc, autoindexingSvc, policiesSvc, policyMatcher, autoindexingSvc.indexEnqueuer, repoStore,
			ConfigIndexingInst.SchedulerInterval,
			background.IndexSchedulerConfig{
				RepositoryProcessDelay: ConfigIndexingInst.RepositoryProcessDelay,
				RepositoryBatchSize:    ConfigIndexingInst.RepositoryBatchSize,
				PolicyBatchSize:        ConfigIndexingInst.PolicyBatchSize,
				InferenceConcurrency:   ConfigIndexingInst.InferenceConcurrency,
			},
		),

		background.NewOnDemandScheduler(
			autoindexingSvc.store,
			autoindexingSvc.indexEnqueuer,
			ConfigIndexingInst.OnDemandSchedulerInterval,
			ConfigIndexingInst.OnDemandBatchsize,
		),
	}
}

func NewDependencyIndexSchedulers(
	observationCtx *observation.Context,
	db database.DB,
	uploadSvc UploadService,
	depsSvc DependenciesService,
	autoindexingSvc *Service,
	repoUpdater RepoUpdaterClient,
) []goroutine.BackgroundRoutine {
	dependencySyncStore := dbworkerstore.New(observationCtx, db.Handle(), background.DependencySyncingJobWorkerStoreOptions)
	dependencyIndexingStore := dbworkerstore.New(observationCtx, db.Handle(), background.DependencyIndexingJobWorkerStoreOptions)

	externalServiceStore := db.ExternalServices()
	repoStore := db.Repos()
	gitserverRepoStore := db.GitserverRepos()

	return []goroutine.BackgroundRoutine{
		background.NewDependencySyncScheduler(
			dependencySyncStore,
			uploadSvc, depsSvc, autoindexingSvc.store, externalServiceStore, workerutil.NewMetrics(observationCtx, "codeintel_dependency_index_processor"),
			ConfigDependencyIndexInst.DependencyIndexerSchedulerPollInterval,
		),
		background.NewDependencyIndexingScheduler(
			dependencyIndexingStore,
			uploadSvc, repoStore, externalServiceStore, gitserverRepoStore, autoindexingSvc.indexEnqueuer, repoUpdater,
			workerutil.NewMetrics(observationCtx, "codeintel_dependency_index_queueing"),
			ConfigDependencyIndexInst.DependencyIndexerSchedulerPollInterval,
			ConfigDependencyIndexInst.DependencyIndexerSchedulerConcurrency,
		),
	}
}

func NewSummaryBuilder(
	observationCtx *observation.Context,
	autoindexingSvc *Service,
	uploadSvc UploadService,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewSummaryBuilder(
			observationCtx,
			autoindexingSvc.store,
			autoindexingSvc.jobSelector,
			uploadSvc,
			SummaryBuilderConfigInst.Interval,
			SummaryBuilderConfigInst.NumRepositoriesToConfigure,
		),
	}
}

package autoindexing

import (
	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var (
	IndexWorkerStoreOptions                 = background.IndexWorkerStoreOptions
	DependencySyncingJobWorkerStoreOptions  = background.DependencySyncingJobWorkerStoreOptions
	DependencyIndexingJobWorkerStoreOptions = background.DependencyIndexingJobWorkerStoreOptions
)

func NewService(
	db database.DB,
	uploadSvc UploadService,
	depsSvc DependenciesService,
	policiesSvc PoliciesService,
	gitserver GitserverClient,
	observationContext *observation.Context,
) *Service {
	store := store.New(db, scopedContext("store", observationContext))
	symbolsClient := symbols.DefaultClient
	repoUpdater := repoupdater.DefaultClient
	inferenceSvc := inference.NewService(db)

	svc := newService(store, uploadSvc, inferenceSvc, repoUpdater, gitserver, symbolsClient, scopedContext("service", observationContext))

	return svc
}

type serviceDependencies struct {
	db                 database.DB
	uploadSvc          UploadService
	depsSvc            DependenciesService
	policiesSvc        PoliciesService
	gitserver          GitserverClient
	observationContext *observation.Context
}

func scopedContext(component string, observationContext *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "autoindexing", component, observationContext)
}

func NewResetters(db database.DB, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationContext)
	indexStore := dbworkerstore.New(db.Handle(), background.IndexWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.New(db.Handle(), background.DependencyIndexingJobWorkerStoreOptions, observationContext)

	return []goroutine.BackgroundRoutine{
		background.NewIndexResetter(ConfigCleanupInst.Interval, indexStore, observationContext.Logger.Scoped("indexResetter", ""), metrics),
		background.NewDependencyIndexResetter(ConfigCleanupInst.Interval, dependencyIndexingStore, observationContext.Logger.Scoped("dependencyIndexResetter", ""), metrics),
	}
}

func NewJanitorJobs(autoindexingSvc *Service, gitserver GitserverClient, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewJanitor(
			ConfigCleanupInst.Interval,
			autoindexingSvc.store, gitserver, glock.NewRealClock(),
			background.JanitorConfig{
				MinimumTimeSinceLastCheck:      ConfigCleanupInst.MinimumTimeSinceLastCheck,
				CommitResolverBatchSize:        ConfigCleanupInst.CommitResolverBatchSize,
				CommitResolverMaximumCommitLag: ConfigCleanupInst.CommitResolverMaximumCommitLag,
				FailedIndexBatchSize:           ConfigCleanupInst.FailedIndexBatchSize,
				FailedIndexMaxAge:              ConfigCleanupInst.FailedIndexMaxAge,
			},
			observationContext,
		),
	}
}

func NewIndexSchedulers(
	uploadSvc UploadService,
	policiesSvc PoliciesService,
	policyMatcher PolicyMatcher,
	autoindexingSvc *Service,
	observationContext *observation.Context,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewScheduler(
			uploadSvc, policiesSvc, policyMatcher, autoindexingSvc.indexEnqueuer,
			ConfigIndexingInst.SchedulerInterval,
			background.IndexSchedulerConfig{
				RepositoryProcessDelay: ConfigIndexingInst.RepositoryProcessDelay,
				RepositoryBatchSize:    ConfigIndexingInst.RepositoryBatchSize,
				PolicyBatchSize:        ConfigIndexingInst.PolicyBatchSize,
				InferenceConcurrency:   ConfigIndexingInst.InferenceConcurrency,
			},
			observationContext,
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
	db database.DB,
	uploadSvc UploadService,
	depsSvc DependenciesService,
	autoindexingSvc *Service,
	repoUpdater RepoUpdaterClient,
	observationContext *observation.Context,
) []goroutine.BackgroundRoutine {
	dependencySyncStore := dbworkerstore.New(db.Handle(), background.DependencySyncingJobWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.New(db.Handle(), background.DependencyIndexingJobWorkerStoreOptions, observationContext)

	externalServiceStore := db.ExternalServices()
	repoStore := db.Repos()
	gitserverRepoStore := db.GitserverRepos()

	return []goroutine.BackgroundRoutine{
		background.NewDependencySyncScheduler(
			dependencySyncStore,
			uploadSvc, depsSvc, autoindexingSvc.store, externalServiceStore, workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor"),
			ConfigDependencyIndexInst.DependencyIndexerSchedulerPollInterval,
		),
		background.NewDependencyIndexingScheduler(
			dependencyIndexingStore,
			uploadSvc, repoStore, externalServiceStore, gitserverRepoStore, autoindexingSvc.indexEnqueuer, repoUpdater,
			workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing"),
			ConfigDependencyIndexInst.DependencyIndexerSchedulerPollInterval,
			ConfigDependencyIndexInst.DependencyIndexerSchedulerConcurrency,
		),
	}
}

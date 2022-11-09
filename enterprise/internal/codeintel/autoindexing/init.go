package autoindexing

import (
	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/memo"
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

// GetService creates or returns an already-initialized autoindexing service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc UploadService,
	depsSvc DependenciesService,
	policiesSvc PoliciesService,
	gitserver GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		uploadSvc,
		depsSvc,
		policiesSvc,
		gitserver,
	})

	return svc
}

type serviceDependencies struct {
	db          database.DB
	uploadSvc   UploadService
	depsSvc     DependenciesService
	policiesSvc PoliciesService
	gitserver   GitserverClient
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	symbolsClient := symbols.DefaultClient
	repoUpdater := repoupdater.DefaultClient
	inferenceSvc := inference.NewService(deps.db)

	svc := newService(store, deps.uploadSvc, inferenceSvc, repoUpdater, deps.gitserver, symbolsClient, scopedContext("service"))

	return svc, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "autoindexing", component)
}

func NewResetters(db database.DB, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationContext)
	indexStore := dbworkerstore.NewWithMetrics(db.Handle(), background.IndexWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.NewWithMetrics(db.Handle(), background.DependencyIndexingJobWorkerStoreOptions, observationContext)

	return []goroutine.BackgroundRoutine{
		background.NewIndexResetter(ConfigCleanupInst.Interval, indexStore, observationContext.Logger.Scoped("indexResetter", ""), metrics),
		background.NewDependencyIndexResetter(ConfigCleanupInst.Interval, dependencyIndexingStore, observationContext.Logger.Scoped("dependencyIndexResetter", ""), metrics),
	}
}

func NewJanitorJobs(autoindexingSvc AutoIndexingService, gitserver GitserverClient) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewJanitor(
			ConfigCleanupInst.Interval,
			autoindexingSvc, gitserver, glock.NewRealClock(),
			background.JanitorConfig{
				MinimumTimeSinceLastCheck:      ConfigCleanupInst.MinimumTimeSinceLastCheck,
				CommitResolverBatchSize:        ConfigCleanupInst.CommitResolverBatchSize,
				CommitResolverMaximumCommitLag: ConfigCleanupInst.CommitResolverMaximumCommitLag,
				FailedIndexBatchSize:           ConfigCleanupInst.FailedIndexBatchSize,
				FailedIndexMaxAge:              ConfigCleanupInst.FailedIndexMaxAge,
			},
		),
	}
}

func NewIndexSchedulers(
	uploadSvc UploadService,
	policiesSvc PoliciesService,
	policyMatcher PolicyMatcher,
	autoindexingSvc AutoIndexingService,
	observationContext *observation.Context,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewScheduler(
			uploadSvc, policiesSvc, policyMatcher, autoindexingSvc,
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
			autoindexingSvc,
			ConfigIndexingInst.OnDemandSchedulerInterval,
			ConfigIndexingInst.OnDemandBatchsize,
		),
	}
}

func NewDependencyIndexSchedulers(
	db database.DB,
	uploadSvc UploadService,
	depsSvc DependenciesService,
	autoindexingSvc AutoIndexingService,
	repoUpdater RepoUpdaterClient,
	observationContext *observation.Context,
) []goroutine.BackgroundRoutine {
	dependencySyncStore := dbworkerstore.NewWithMetrics(db.Handle(), background.DependencySyncingJobWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.NewWithMetrics(db.Handle(), background.DependencyIndexingJobWorkerStoreOptions, observationContext)

	externalServiceStore := db.ExternalServices()
	repoStore := db.Repos()
	gitserverRepoStore := db.GitserverRepos()

	return []goroutine.BackgroundRoutine{
		background.NewDependencySyncScheduler(
			dependencySyncStore,
			uploadSvc, depsSvc, autoindexingSvc, externalServiceStore, workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor"),
			ConfigDependencyIndexInst.DependencyIndexerSchedulerPollInterval,
		),
		background.NewDependencyIndexingScheduler(
			dependencyIndexingStore,
			uploadSvc, repoStore, externalServiceStore, gitserverRepoStore, autoindexingSvc, repoUpdater,
			workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing"),
			ConfigDependencyIndexInst.DependencyIndexerSchedulerPollInterval,
			ConfigDependencyIndexInst.DependencyIndexerSchedulerConcurrency,
		),
	}
}

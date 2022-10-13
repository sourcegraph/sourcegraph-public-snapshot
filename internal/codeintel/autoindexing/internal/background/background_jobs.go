package background

import (
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type BackgroundJob interface {
	NewDependencyIndexingScheduler(pollInterval time.Duration, numHandlers int) *workerutil.Worker
	NewDependencySyncScheduler(pollInterval time.Duration) *workerutil.Worker
	NewDependencyIndexResetter(interval time.Duration) *dbworker.Resetter
	NewIndexResetter(interval time.Duration) *dbworker.Resetter
	NewOnDemandScheduler(interval time.Duration, batchSize int) goroutine.BackgroundRoutine
	NewScheduler(interval time.Duration, repositoryProcessDelay time.Duration, repositoryBatchSize int, policyBatchSize int) goroutine.BackgroundRoutine
	NewJanitor(
		interval time.Duration,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
	) goroutine.BackgroundRoutine

	SetService(service AutoIndexingService)
	WorkerutilStore() dbworkerstore.Store
	DependencySyncStore() dbworkerstore.Store
	DependencyIndexingStore() dbworkerstore.Store
}

type backgroundJob struct {
	uploadSvc       UploadService
	depsSvc         DependenciesService
	policiesSvc     PoliciesService
	autoindexingSvc AutoIndexingService

	policyMatcher   PolicyMatcher
	repoUpdater     RepoUpdaterClient
	gitserverClient GitserverClient

	repoStore               ReposStore
	workerutilStore         dbworkerstore.Store
	gitserverRepoStore      GitserverRepoStore
	dependencySyncStore     dbworkerstore.Store
	externalServiceStore    ExternalServiceStore
	dependencyIndexingStore dbworkerstore.Store

	operations *operations
	clock      glock.Clock
	logger     log.Logger

	metrics                *resetterMetrics
	janitorMetrics         *janitorMetrics
	depencencySyncMetrics  workerutil.WorkerObservability
	depencencyIndexMetrics workerutil.WorkerObservability
}

func New(
	db database.DB,
	uploadSvc UploadService,
	depsSvc DependenciesService,
	policiesSvc PoliciesService,
	policyMatcher PolicyMatcher,
	gitserverClient GitserverClient,
	repoUpdater RepoUpdaterClient,
	observationContext *observation.Context,
) BackgroundJob {
	repoStore := db.Repos()
	gitserverRepoStore := db.GitserverRepos()
	externalServiceStore := db.ExternalServices()
	workerutilStore := dbworkerstore.NewWithMetrics(db.Handle(), indexWorkerStoreOptions, observationContext)
	dependencySyncStore := dbworkerstore.NewWithMetrics(db.Handle(), dependencySyncingJobWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.NewWithMetrics(db.Handle(), dependencyIndexingJobWorkerStoreOptions, observationContext)

	return &backgroundJob{
		uploadSvc:   uploadSvc,
		depsSvc:     depsSvc,
		policiesSvc: policiesSvc,

		policyMatcher:   policyMatcher,
		repoUpdater:     repoUpdater,
		gitserverClient: gitserverClient,

		repoStore:               repoStore,
		workerutilStore:         workerutilStore,
		gitserverRepoStore:      gitserverRepoStore,
		dependencySyncStore:     dependencySyncStore,
		externalServiceStore:    externalServiceStore,
		dependencyIndexingStore: dependencyIndexingStore,

		operations: newOperations(observationContext),
		clock:      glock.NewRealClock(),
		logger:     observationContext.Logger,

		metrics:                newResetterMetrics(observationContext),
		janitorMetrics:         newJanitorMetrics(observationContext),
		depencencySyncMetrics:  workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor"),
		depencencyIndexMetrics: workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing"),
	}
}

func (b *backgroundJob) SetService(service AutoIndexingService) {
	b.autoindexingSvc = service
}

func (b backgroundJob) WorkerutilStore() dbworkerstore.Store     { return b.workerutilStore }
func (b backgroundJob) DependencySyncStore() dbworkerstore.Store { return b.dependencySyncStore }
func (b backgroundJob) DependencyIndexingStore() dbworkerstore.Store {
	return b.dependencyIndexingStore
}

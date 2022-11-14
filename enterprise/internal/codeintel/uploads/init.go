package uploads

import (
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesEnterprise "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/enterprise"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// GetService creates or returns an already-initialized uploads service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	gsc GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		codeIntelDB,
		gsc,
	})

	return svc
}

type serviceDependencies struct {
	db          database.DB
	codeIntelDB codeintelshared.CodeIntelDB
	gsc         GitserverClient
}

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	repoStore := backend.NewRepos(scopedContext("repos").Logger, deps.db, gitserver.NewClient(deps.db))
	lsifStore := lsifstore.New(deps.codeIntelDB, scopedContext("lsifstore"))
	policyMatcher := policiesEnterprise.NewMatcher(deps.gsc, policiesEnterprise.RetentionExtractor, true, false)
	locker := locker.NewWith(deps.db, "codeintel")

	svc := newService(
		store,
		repoStore,
		lsifStore,
		deps.gsc,
		nil, // written in circular fashion
		policyMatcher,
		locker,
		scopedContext("service"),
	)
	svc.policySvc = policies.GetService(deps.db, svc, deps.gsc)

	return svc, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "uploads", component)
}

func NewUploadProcessorJob(
	uploadSvc UploadService,
	db database.DB,
	uploadStore uploadstore.Store,
	workerConcurrency int,
	workerBudget int64,
	workerPollInterval time.Duration,
	maximumRuntimePerJob time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	uploadsProcessorStore := dbworkerstore.NewWithMetrics(db.Handle(), store.UploadWorkerStoreOptions, observationContext)
	return background.NewUploadProcessorWorker(
		uploadSvc,
		uploadsProcessorStore,
		uploadStore,
		workerConcurrency,
		workerBudget,
		workerPollInterval,
		maximumRuntimePerJob,
		observationContext,
	)
}

func NewCommittedAtBackfillerJob(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCommittedAtBackfiller(
			uploadSvc,
			ConfigCommittedAtBackfillInst.Interval,
			ConfigCommittedAtBackfillInst.BatchSize,
		),
	}
}

func NewJanitor(uploadSvc UploadService, gitserverClient GitserverClient, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewJanitor(
			uploadSvc,
			gitserverClient,
			ConfigJanitorInst.Interval,
			background.JanitorConfig{
				UploadTimeout:                  ConfigJanitorInst.UploadTimeout,
				AuditLogMaxAge:                 ConfigJanitorInst.AuditLogMaxAge,
				MinimumTimeSinceLastCheck:      ConfigJanitorInst.MinimumTimeSinceLastCheck,
				CommitResolverBatchSize:        ConfigJanitorInst.CommitResolverBatchSize,
				CommitResolverMaximumCommitLag: ConfigJanitorInst.CommitResolverMaximumCommitLag,
			},
			glock.NewRealClock(),
			observationContext.Logger,
			background.NewJanitorMetrics(observationContext),
		),
	}
}

func NewReconciler(uploadSvc UploadService, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewReconciler(uploadSvc, ConfigJanitorInst.Interval, ConfigJanitorInst.ReconcilerBatchSize, observationContext),
	}
}

func NewResetters(db database.DB, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationContext)
	uploadsResetterStore := dbworkerstore.NewWithMetrics(db.Handle(), store.UploadWorkerStoreOptions, observationContext)

	return []goroutine.BackgroundRoutine{
		background.NewUploadResetter(uploadsResetterStore, ConfigJanitorInst.Interval, observationContext.Logger.Scoped("uploadResetter", ""), metrics),
	}
}

func NewCommitGraphUpdater(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCommitGraphUpdater(
			uploadSvc,
			ConfigCommitGraphInst.CommitGraphUpdateTaskInterval,
			ConfigCommitGraphInst.MaxAgeForNonStaleBranches,
			ConfigCommitGraphInst.MaxAgeForNonStaleTags,
		),
	}
}

func NewExpirationTasks(uploadSvc UploadService, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewUploadExpirer(
			uploadSvc,
			ConfigExpirationInst.ExpirerInterval,
			background.ExpirerConfig{
				RepositoryProcessDelay: ConfigExpirationInst.RepositoryProcessDelay,
				RepositoryBatchSize:    ConfigExpirationInst.RepositoryBatchSize,
				UploadProcessDelay:     ConfigExpirationInst.UploadProcessDelay,
				UploadBatchSize:        ConfigExpirationInst.UploadBatchSize,
				CommitBatchSize:        ConfigExpirationInst.CommitBatchSize,
				PolicyBatchSize:        ConfigExpirationInst.PolicyBatchSize,
			},
			observationContext,
		),
	}
}

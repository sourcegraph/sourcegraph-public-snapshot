package background

import (
	"time"

	"github.com/derision-test/glock"
	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type BackgroundJob interface {
	NewCommittedAtBackfiller(interval time.Duration, batchSize int) goroutine.BackgroundRoutine
	NewJanitor(
		interval time.Duration,
		uploadTimeout time.Duration,
		auditLogMaxAge time.Duration,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
	) goroutine.BackgroundRoutine
	NewUploadResetter(interval time.Duration) *dbworker.Resetter
	NewCommitGraphUpdater(interval time.Duration, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) goroutine.BackgroundRoutine
	NewWorker(
		uploadStore uploadstore.Store,
		workerConcurrency int,
		workerBudget int64,
		workerPollInterval time.Duration,
		maximumRuntimePerJob time.Duration,
	) *workerutil.Worker

	SetUploadsService(s UploadService)
}

type backgroundJob struct {
	uploadSvc       UploadService
	gitserverClient GitserverClient

	// workerutilStore dbworkerstore.Store

	clock           glock.Clock
	logger          logger.Logger
	janitorMetrics  *janitorMetrics
	resetterMetrics *resetterMetrics
	workerMetrics   workerutil.WorkerObservability
	operations      *operations
}

func New(db database.DB, gsc GitserverClient, observationContext *observation.Context) BackgroundJob {
	// workerutilStore := dbworkerstore.NewWithMetrics(db.Handle(), uploadWorkerStoreOptions, observationContext)

	// // TODO - move this to metric reporter?
	// dbworker.InitPrometheusMetric(observationContext, workerutilStore, "codeintel", "upload", nil)

	return &backgroundJob{
		gitserverClient: gsc,
		// workerutilStore: workerutilStore,

		clock:           glock.NewRealClock(),
		logger:          observationContext.Logger,
		janitorMetrics:  newJanitorMetrics(observationContext),
		resetterMetrics: newResetterMetrics(observationContext),
		workerMetrics:   workerutil.NewMetrics(observationContext, "codeintel_upload_processor", workerutil.WithSampler(func(job workerutil.Record) bool { return true })),
		operations:      newOperations(observationContext),
	}
}

func (b *backgroundJob) SetUploadsService(s UploadService) {
	b.uploadSvc = s
}

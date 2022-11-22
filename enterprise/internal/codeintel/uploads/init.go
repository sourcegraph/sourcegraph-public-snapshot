package uploads

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/derision-test/glock"
	"google.golang.org/api/option"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	policiesEnterprise "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/enterprise"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/env"
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

var (
	bucketName                   = env.Get("CODEINTEL_UPLOADS_RANKING_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
	rankingGraphKey              = env.Get("CODEINTEL_UPLOADS_RANKING_GRAPH_KEY", "dev", "An identifier of the graph export. Change to start a new export in the configured bucket.")
	rankingGraphBatchSize        = env.MustGetInt("CODEINTEL_UPLOADS_RANKING_GRAPH_BATCH_SIZE", 16, "How many uploads to process at once.")
	rankingGraphDeleteBatchSize  = env.MustGetInt("CODEINTEL_UPLOADS_RANKING_GRAPH_DELETE_BATCH_SIZE", 32, "How many stale uploads to delete at once.")
	rankingBucketCredentialsFile = env.Get("CODEINTEL_UPLOADS_RANKING_GOOGLE_APPLICATION_CREDENTIALS_FILE", "", "The path to a service account key file with access to GCS.")
)

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	repoStore := backend.NewRepos(scopedContext("repos").Logger, deps.db, gitserver.NewClient(deps.db))
	lsifStore := lsifstore.New(deps.codeIntelDB, scopedContext("lsifstore"))
	policyMatcher := policiesEnterprise.NewMatcher(deps.gsc, policiesEnterprise.RetentionExtractor, true, false)
	locker := locker.NewWith(deps.db, "codeintel")

	rankingBucket := func() *storage.BucketHandle {
		if rankingBucketCredentialsFile == "" && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
			return nil
		}

		var opts []option.ClientOption
		if rankingBucketCredentialsFile != "" {
			opts = append(opts, option.WithCredentialsFile(rankingBucketCredentialsFile))
		}

		client, err := storage.NewClient(context.Background(), opts...)
		if err != nil {
			log.Scoped("codenav", "").Error("failed to create storage client", log.Error(err))
			return nil
		}

		return client.Bucket(bucketName)
	}()

	svc := newService(
		store,
		repoStore,
		lsifStore,
		deps.gsc,
		rankingBucket,
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
	uploadSvc *Service,
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
		uploadSvc.store,
		uploadSvc.lsifstore,
		uploadSvc.gitserverClient,
		uploadSvc.repoStore,
		uploadsProcessorStore,
		uploadStore,
		workerConcurrency,
		workerBudget,
		workerPollInterval,
		maximumRuntimePerJob,
		observationContext,
	)
}

func NewCommittedAtBackfillerJob(uploadSvc *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCommittedAtBackfiller(
			uploadSvc.store,
			uploadSvc.gitserverClient,
			ConfigCommittedAtBackfillInst.Interval,
			ConfigCommittedAtBackfillInst.BatchSize,
		),
	}
}

func NewJanitor(uploadSvc *Service, gitserverClient GitserverClient, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewJanitor(
			uploadSvc.store,
			uploadSvc.lsifstore,
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

func NewReconciler(uploadSvc *Service, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewReconciler(uploadSvc.store, uploadSvc.lsifstore, ConfigJanitorInst.Interval, ConfigJanitorInst.ReconcilerBatchSize, observationContext),
	}
}

func NewResetters(db database.DB, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationContext)
	uploadsResetterStore := dbworkerstore.NewWithMetrics(db.Handle(), store.UploadWorkerStoreOptions, observationContext)

	return []goroutine.BackgroundRoutine{
		background.NewUploadResetter(uploadsResetterStore, ConfigJanitorInst.Interval, observationContext.Logger.Scoped("uploadResetter", ""), metrics),
	}
}

func NewCommitGraphUpdater(uploadSvc *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCommitGraphUpdater(
			uploadSvc.store,
			uploadSvc.locker,
			uploadSvc.gitserverClient,
			ConfigCommitGraphInst.CommitGraphUpdateTaskInterval,
			ConfigCommitGraphInst.MaxAgeForNonStaleBranches,
			ConfigCommitGraphInst.MaxAgeForNonStaleTags,
		),
	}
}

func NewExpirationTasks(uploadSvc *Service, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewUploadExpirer(
			uploadSvc.store,
			uploadSvc.policySvc,
			uploadSvc.policyMatcher,
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

func NewGraphExporters(uploadSvc *Service, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRankingGraphExporter(
			uploadSvc,
			ConfigExportInst.NumRankingRoutines,
			ConfigExportInst.RankingInterval,
			observationContext,
		),
	}
}

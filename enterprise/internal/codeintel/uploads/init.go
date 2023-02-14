package uploads

import (
	"context"
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
	uploadsstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	gsc GitserverClient,
) *Service {
	store := uploadsstore.New(scopedContext("uploadsstore", observationCtx), db)
	repoStore := backend.NewRepos(scopedContext("repos", observationCtx).Logger, db, gitserver.NewClient())
	lsifStore := lsifstore.New(scopedContext("lsifstore", observationCtx), codeIntelDB)
	policyMatcher := policiesEnterprise.NewMatcher(gsc, policiesEnterprise.RetentionExtractor, true, false)
	ciLocker := locker.NewWith(db, "codeintel")

	rankingBucket := func() *storage.BucketHandle {
		if rankingBucketCredentialsFile == "" {
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
		scopedContext("service", observationCtx),
		store,
		repoStore,
		lsifStore,
		gsc,
		rankingBucket,
		nil, // written in circular fashion
		policyMatcher,
		ciLocker,
	)
	svc.policySvc = policies.NewService(observationCtx, db, svc, gsc)

	return svc
}

var (
	bucketName                   = env.Get("CODEINTEL_UPLOADS_RANKING_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
	rankingGraphKey              = env.Get("CODEINTEL_UPLOADS_RANKING_GRAPH_KEY", "dev", "An identifier of the graph export. Change to start a new export in the configured bucket.")
	rankingGraphBatchSize        = env.MustGetInt("CODEINTEL_UPLOADS_RANKING_GRAPH_BATCH_SIZE", 16, "How many uploads to process at once.")
	rankingGraphDeleteBatchSize  = env.MustGetInt("CODEINTEL_UPLOADS_RANKING_GRAPH_DELETE_BATCH_SIZE", 32, "How many stale uploads to delete at once.")
	rankingBucketCredentialsFile = env.Get("CODEINTEL_UPLOADS_RANKING_GOOGLE_APPLICATION_CREDENTIALS_FILE", "", "The path to a service account key file with access to GCS.")
)

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "uploads", component, parent)
}

func NewUploadProcessorJob(
	observationCtx *observation.Context,
	uploadSvc *Service,
	db database.DB,
	uploadStore uploadstore.Store,
	workerConcurrency int,
	workerBudget int64,
	workerPollInterval time.Duration,
	maximumRuntimePerJob time.Duration,
) goroutine.BackgroundRoutine {
	uploadsProcessorStore := dbworkerstore.New(observationCtx, db.Handle(), uploadsstore.UploadWorkerStoreOptions)

	dbworker.InitPrometheusMetric(observationCtx, uploadsProcessorStore, "codeintel", "upload", nil)

	return background.NewUploadProcessorWorker(
		observationCtx,
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

func NewJanitor(observationCtx *observation.Context, uploadSvc *Service, gitserverClient GitserverClient) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewJanitor(
			uploadSvc.store,
			uploadSvc.lsifstore,
			gitserverClient,
			ConfigJanitorInst.Interval,
			background.JanitorConfig{
				UploadTimeout:                  ConfigJanitorInst.UploadTimeout,
				AuditLogMaxAge:                 ConfigJanitorInst.AuditLogMaxAge,
				UnreferencedDocumentBatchSize:  ConfigJanitorInst.UnreferencedDocumentBatchSize,
				UnreferencedDocumentMaxAge:     ConfigJanitorInst.UnreferencedDocumentMaxAge,
				MinimumTimeSinceLastCheck:      ConfigJanitorInst.MinimumTimeSinceLastCheck,
				CommitResolverBatchSize:        ConfigJanitorInst.CommitResolverBatchSize,
				CommitResolverMaximumCommitLag: ConfigJanitorInst.CommitResolverMaximumCommitLag,
			},
			glock.NewRealClock(),
			observationCtx.Logger,
			background.NewJanitorMetrics(observationCtx),
		),
	}
}

func NewReconciler(observationCtx *observation.Context, uploadSvc *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewReconciler(observationCtx, uploadSvc.store, uploadSvc.lsifstore, ConfigJanitorInst.Interval, ConfigJanitorInst.ReconcilerBatchSize),
	}
}

func NewResetters(observationCtx *observation.Context, db database.DB) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationCtx)
	uploadsResetterStore := dbworkerstore.New(observationCtx, db.Handle(), uploadsstore.UploadWorkerStoreOptions)

	return []goroutine.BackgroundRoutine{
		background.NewUploadResetter(observationCtx.Logger, uploadsResetterStore, ConfigJanitorInst.Interval, metrics),
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

func NewExpirationTasks(observationCtx *observation.Context, uploadSvc *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewUploadExpirer(
			observationCtx,
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
		),
	}
}

func NewGraphExporters(observationCtx *observation.Context, uploadSvc *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRankingGraphExporter(
			observationCtx,
			uploadSvc,
			ConfigExportInst.NumRankingRoutines,
			ConfigExportInst.RankingInterval,
		),
	}
}

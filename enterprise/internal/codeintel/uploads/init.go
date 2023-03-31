package uploads

import (
	"context"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
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
	gitserverClient gitserver.Client,
) *Service {
	store := uploadsstore.New(scopedContext("uploadsstore", observationCtx), db)
	repoStore := backend.NewRepos(scopedContext("repos", observationCtx).Logger, db, gitserverClient)
	lsifStore := lsifstore.New(scopedContext("lsifstore", observationCtx), codeIntelDB)
	policyMatcher := policies.NewMatcher(gitserverClient, policies.RetentionExtractor, true, false)
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
		gitserverClient,
		rankingBucket,
		nil, // written in circular fashion
		policyMatcher,
		ciLocker,
	)
	svc.policySvc = policies.NewService(observationCtx, db, svc, gitserverClient)

	return svc
}

var (
	bucketName                   = env.Get("CODEINTEL_UPLOADS_RANKING_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
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

func NewCommittedAtBackfillerJob(uploadSvc *Service, gitserverClient gitserver.Client) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCommittedAtBackfiller(
			uploadSvc.store,
			gitserverClient,
			ConfigCommittedAtBackfillInst.Interval,
			ConfigCommittedAtBackfillInst.BatchSize,
		),
	}
}

func NewJanitor(observationCtx *observation.Context, uploadSvc *Service, gitserverClient gitserver.Client) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewDeletedRepositoryJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			observationCtx,
		),

		background.NewUnknownCommitJanitor(
			uploadSvc.store,
			gitserverClient,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.CommitResolverBatchSize,
			ConfigJanitorInst.MinimumTimeSinceLastCheck,
			ConfigJanitorInst.CommitResolverMaximumCommitLag,
			observationCtx,
		),

		background.NewAbandonedUploadJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.UploadTimeout,
			observationCtx,
		),

		background.NewExpiredUploadJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			observationCtx,
		),

		background.NewExpiredUploadTraversalJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			observationCtx,
		),

		background.NewHardDeleter(
			uploadSvc.store,
			uploadSvc.lsifstore,
			ConfigJanitorInst.Interval,
			observationCtx,
		),

		background.NewAuditLogJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.AuditLogMaxAge,
			observationCtx,
		),

		background.NewSCIPExpirationTask(
			uploadSvc.lsifstore,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.UnreferencedDocumentBatchSize,
			ConfigJanitorInst.UnreferencedDocumentMaxAge,
			observationCtx,
		),

		background.NewUnknownRepositoryJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			observationCtx,
		),

		background.NewUnknownCommitJanitor2(
			uploadSvc.store,
			gitserverClient,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.CommitResolverBatchSize,
			ConfigJanitorInst.MinimumTimeSinceLastCheck,
			ConfigJanitorInst.CommitResolverMaximumCommitLag,
			observationCtx,
		),

		background.NewExpiredRecordJanitor(
			uploadSvc.store,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.FailedIndexBatchSize,
			ConfigJanitorInst.FailedIndexMaxAge,
			observationCtx,
		),
	}
}

func NewReconciler(observationCtx *observation.Context, uploadSvc *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewFrontendDBReconciler(
			uploadSvc.store,
			uploadSvc.lsifstore,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.ReconcilerBatchSize,
			observationCtx,
		),

		background.NewCodeIntelDBReconciler(
			uploadSvc.store,
			uploadSvc.lsifstore,
			ConfigJanitorInst.Interval,
			ConfigJanitorInst.ReconcilerBatchSize,
			observationCtx,
		),
	}
}

func NewResetters(observationCtx *observation.Context, db database.DB) []goroutine.BackgroundRoutine {
	metrics := background.NewResetterMetrics(observationCtx)
	uploadsResetterStore := dbworkerstore.New(observationCtx, db.Handle(), uploadsstore.UploadWorkerStoreOptions)

	return []goroutine.BackgroundRoutine{
		background.NewUploadResetter(observationCtx.Logger, uploadsResetterStore, ConfigJanitorInst.Interval, metrics),
	}
}

func NewCommitGraphUpdater(uploadSvc *Service, gitserverClient gitserver.Client) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewCommitGraphUpdater(
			uploadSvc.store,
			uploadSvc.locker,
			gitserverClient,
			ConfigCommitGraphInst.CommitGraphUpdateTaskInterval,
			ConfigCommitGraphInst.MaxAgeForNonStaleBranches,
			ConfigCommitGraphInst.MaxAgeForNonStaleTags,
		),
	}
}

func NewExpirationTasks(observationCtx *observation.Context, uploadSvc *Service, repoStore database.RepoStore) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewUploadExpirer(
			observationCtx,
			uploadSvc.store,
			repoStore,
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

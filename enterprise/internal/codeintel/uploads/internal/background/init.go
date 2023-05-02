package background

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/backfiller"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/expirer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/processor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	uploadsstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewUploadProcessorJob(
	observationCtx *observation.Context,
	store uploadsstore.Store,
	lsifstore lsifstore.Store,
	repoStore processor.RepoStore,
	gitserverClient gitserver.Client,
	db database.DB,
	uploadStore uploadstore.Store,
	config *processor.Config,
) []goroutine.BackgroundRoutine {
	metrics := processor.NewResetterMetrics(observationCtx)
	uploadsProcessorStore := dbworkerstore.New(observationCtx, db.Handle(), uploadsstore.UploadWorkerStoreOptions)
	uploadsResetterStore := dbworkerstore.New(observationCtx.Clone(observation.Honeycomb(nil)), db.Handle(), uploadsstore.UploadWorkerStoreOptions)
	dbworker.InitPrometheusMetric(observationCtx, uploadsProcessorStore, "codeintel", "upload", nil)

	return []goroutine.BackgroundRoutine{
		processor.NewUploadProcessorWorker(
			observationCtx,
			store,
			lsifstore,
			gitserverClient,
			repoStore,
			uploadsProcessorStore,
			uploadStore,
			config,
		),
		processor.NewUploadResetter(observationCtx.Logger, uploadsResetterStore, config, metrics),
	}
}

func NewCommittedAtBackfillerJob(
	store uploadsstore.Store,
	gitserverClient gitserver.Client,
	config *backfiller.Config,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backfiller.NewCommittedAtBackfiller(
			store,
			gitserverClient,
			config,
		),
	}
}

func NewJanitor(
	observationCtx *observation.Context,
	store uploadsstore.Store,
	lsifstore lsifstore.Store,
	gitserverClient gitserver.Client,
	config *janitor.Config,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		janitor.NewDeletedRepositoryJanitor(store, config, observationCtx),
		janitor.NewUnknownCommitJanitor(store, gitserverClient, config, observationCtx),
		janitor.NewAbandonedUploadJanitor(store, config, observationCtx),
		janitor.NewExpiredUploadJanitor(store, config, observationCtx),
		janitor.NewExpiredUploadTraversalJanitor(store, config, observationCtx),
		janitor.NewHardDeleter(store, lsifstore, config, observationCtx),
		janitor.NewAuditLogJanitor(store, config, observationCtx),
		janitor.NewSCIPExpirationTask(lsifstore, config, observationCtx),
		janitor.NewUnknownRepositoryJanitor(store, config, observationCtx),
		janitor.NewUnknownCommitJanitor2(store, gitserverClient, config, observationCtx),
		janitor.NewExpiredRecordJanitor(store, config, observationCtx),
		janitor.NewFrontendDBReconciler(store, lsifstore, config, observationCtx),
		janitor.NewCodeIntelDBReconciler(store, lsifstore, config, observationCtx),
	}
}

func NewCommitGraphUpdater(
	store uploadsstore.Store,
	gitserverClient gitserver.Client,
	config *commitgraph.Config,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		commitgraph.NewCommitGraphUpdater(
			store,
			gitserverClient,
			config,
		),
	}
}

func NewExpirationTasks(
	observationCtx *observation.Context,
	store uploadsstore.Store,
	policySvc expirer.PolicyService,
	gitserverClient gitserver.Client,
	repoStore database.RepoStore,
	config *expirer.Config,
) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		expirer.NewUploadExpirer(
			observationCtx,
			store,
			repoStore,
			policySvc,
			gitserverClient,
			config,
		),
	}
}

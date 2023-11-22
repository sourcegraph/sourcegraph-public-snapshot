package uploads

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background/backfiller"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background/expirer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background/janitor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/background/processor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	uploadsstore "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
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

	svc := newService(
		scopedContext("service", observationCtx),
		store,
		repoStore,
		lsifStore,
		gitserverClient,
	)

	return svc
}

var (
	bucketName                   = env.Get("CODEINTEL_UPLOADS_RANKING_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
	rankingBucketCredentialsFile = env.Get("CODEINTEL_UPLOADS_RANKING_GOOGLE_APPLICATION_CREDENTIALS_FILE", "", "The path to a service account key file with access to GCS.")
)

var (
	BackfillerConfigInst  = &backfiller.Config{}
	CommitGraphConfigInst = &commitgraph.Config{}
	ExpirerConfigInst     = &expirer.Config{}
	JanitorConfigInst     = &janitor.Config{}
	ProcessorConfigInst   = &processor.Config{}
)

func NewUploadProcessorJob(
	observationCtx *observation.Context,
	uploadSvc *Service,
	db database.DB,
	uploadStore uploadstore.Store,
	workerConcurrency int,
	workerBudget int64,
	workerPollInterval time.Duration,
	maximumRuntimePerJob time.Duration,
) []goroutine.BackgroundRoutine {
	ProcessorConfigInst.WorkerConcurrency = workerConcurrency
	ProcessorConfigInst.WorkerBudget = workerBudget
	ProcessorConfigInst.WorkerPollInterval = workerPollInterval
	ProcessorConfigInst.MaximumRuntimePerJob = maximumRuntimePerJob

	return background.NewUploadProcessorJob(
		scopedContext("processor", observationCtx),
		uploadSvc.store,
		uploadSvc.lsifstore,
		uploadSvc.repoStore,
		uploadSvc.gitserverClient,
		db,
		uploadStore,
		ProcessorConfigInst,
	)
}

func NewCommittedAtBackfillerJob(
	uploadSvc *Service,
	gitserverClient gitserver.Client,
) []goroutine.BackgroundRoutine {
	return background.NewCommittedAtBackfillerJob(
		// TODO - context
		uploadSvc.store,
		gitserverClient,
		BackfillerConfigInst,
	)
}

func NewJanitor(
	observationCtx *observation.Context,
	uploadSvc *Service,
	gitserverClient gitserver.Client,
) []goroutine.BackgroundRoutine {
	return background.NewJanitor(
		scopedContext("janitor", observationCtx),
		uploadSvc.store,
		uploadSvc.lsifstore,
		gitserverClient,
		JanitorConfigInst,
	)
}

func NewCommitGraphUpdater(
	uploadSvc *Service,
	gitserverClient gitserver.Client,
) []goroutine.BackgroundRoutine {
	return background.NewCommitGraphUpdater(
		// TODO - context
		uploadSvc.store,
		gitserverClient,
		CommitGraphConfigInst,
	)
}

func NewExpirationTasks(
	observationCtx *observation.Context,
	uploadSvc *Service,
	policySvc expirer.PolicyService,
	repoStore database.RepoStore,
) []goroutine.BackgroundRoutine {
	return background.NewExpirationTasks(
		scopedContext("expiration", observationCtx),
		uploadSvc.store,
		policySvc,
		uploadSvc.gitserverClient,
		repoStore,
		ExpirerConfigInst,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "uploads", component, parent)
}

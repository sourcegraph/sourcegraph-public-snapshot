package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/codeintel/indexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type indexingJob struct{}

func NewIndexingJob() job.Job {
	return &indexingJob{}
}

func (j *indexingJob) Description() string {
	return ""
}

func (j *indexingJob) Config() []env.Config {
	return []env.Config{indexingConfigInst}
}

func (j *indexingJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	logger = logger.Scoped("routines", "indexing job routines")
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "indexing job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize stores
	rawDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, rawDB)
	repoStore := database.ReposWith(logger, db)
	extSvcStore := db.ExternalServices()
	gitserverRepoStore := db.GitserverRepos()

	rawCodeIntelDB, err := codeintel.InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}
	codeIntelDB := database.NewDB(logger, rawCodeIntelDB)

	// Initialize clients
	repoUpdaterClient := codeintel.InitRepoUpdaterClient()
	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	// Get services
	uploadSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	depsSvc := dependencies.GetService(db)
	autoindexingSvc := autoindexing.GetService(db, uploadSvc, gitserverClient, repoUpdaterClient)
	dependencySyncStore := autoindexingSvc.DependencySyncStore()
	dependencyIndexingStore := autoindexingSvc.DependencyIndexingStore()

	// Initialize metrics
	dbworker.InitPrometheusMetric(observationContext, dependencySyncStore, "codeintel", "dependency_index", nil)
	syncMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor")
	queueingMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing")

	routines := []goroutine.BackgroundRoutine{
		indexing.NewDependencySyncScheduler(uploadSvc, depsSvc, autoindexingSvc, dependencySyncStore, extSvcStore, syncMetrics, observationContext),
		indexing.NewDependencyIndexingScheduler(uploadSvc, repoStore, dependencyIndexingStore, extSvcStore, gitserverRepoStore, repoUpdaterClient, autoindexingSvc, indexingConfigInst.DependencyIndexerSchedulerPollInterval, indexingConfigInst.DependencyIndexerSchedulerConcurrency, queueingMetrics),
	}

	return routines, nil
}

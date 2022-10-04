package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	bkgdependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/dependencies"
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

type autoindexingDependencyScheduler struct{}

func NewAutoindexingDependencySchedulerJob() job.Job {
	return &autoindexingDependencyScheduler{}
}

func (j *autoindexingDependencyScheduler) Description() string {
	return ""
}

func (j *autoindexingDependencyScheduler) Config() []env.Config {
	return []env.Config{
		bkgdependencies.ConfigInst,
	}
}

func (j *autoindexingDependencyScheduler) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel autoindexing dependency scheduling routines"),
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
	codeintelDB := database.NewDB(logger, rawCodeIntelDB)

	// Initialize necessary clients
	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}
	repoUpdater := codeintel.InitRepoUpdaterClient()

	// Initialize services
	uploadSvc := uploads.GetService(db, codeintelDB, gitserverClient)
	autoindexingSvc := autoindexing.GetService(db, uploadSvc, gitserverClient, repoUpdater)
	depsSvc := dependencies.GetService(db)
	dependencySyncStore := autoindexingSvc.DependencySyncStore()
	dependencyIndexingStore := autoindexingSvc.DependencyIndexingStore()

	syncMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor")
	queueingMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing")

	// Initialize metrics
	dbworker.InitPrometheusMetric(observationContext, dependencySyncStore, "codeintel", "dependency_index", nil)

	// Initialize services
	return []goroutine.BackgroundRoutine{
		bkgdependencies.NewDependencySyncScheduler(uploadSvc, depsSvc, autoindexingSvc, dependencySyncStore, extSvcStore, syncMetrics, observationContext),
		bkgdependencies.NewDependencyIndexingScheduler(uploadSvc, repoStore, dependencyIndexingStore, extSvcStore, gitserverRepoStore, repoUpdater, autoindexingSvc, queueingMetrics),
	}, nil
}

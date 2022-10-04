package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/dependencies"
	bkgdependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
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
	// observationContext := &observation.Context{
	// 	Logger:     logger.Scoped("routines", "codeintel autoindexing dependency scheduling routines"),
	// 	Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
	// 	Registerer: prometheus.DefaultRegisterer,
	// }

	// Initialize stores
	rawDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, rawDB)

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
	autoIndexingSvc := autoindexing.GetService(db, uploadSvc, gitserverClient, repoUpdater)
	// dependencySyncStore := autoIndexingSvc.DependencySyncStore()
	// dependencyIndexingStore := autoindexingSvc.DependencyIndexingStore()

	// syncMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor")
	// queueingMetrics := workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing")

	// TODO - move these to metrics reporter job
	// Initialize metrics
	// dbworker.InitPrometheusMetric(observationContext, dependencySyncStore, "codeintel", "dependency_index", nil)

	return dependencies.NewSchedulers(autoIndexingSvc), nil
}

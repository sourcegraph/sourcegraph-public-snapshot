package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type autoindexingScheduler struct{}

func NewAutoindexingSchedulerJob() job.Job {
	return &autoindexingScheduler{}
}

func (j *autoindexingScheduler) Description() string {
	return ""
}

func (j *autoindexingScheduler) Config() []env.Config {
	return []env.Config{
		scheduler.ConfigInst,
	}
}

func (j *autoindexingScheduler) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	// observationContext := &observation.Context{
	// 	Logger:     logger.Scoped("routines", "codeintel autoindexing scheduling routines"),
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
	// TODO
	// policyMatcher := policiesEnterprise.NewMatcher(gitserverClient, policiesEnterprise.IndexingExtractor, false, true)

	// Initialize services
	uploadSvc := uploads.GetService(db, codeintelDB, gitserverClient)
	depsSvc := dependencies.GetService(db)
	policySvc := policies.GetService(db, uploadSvc, gitserverClient)
	autoIndexingSvc := autoindexing.GetService(db, uploadSvc, depsSvc, policySvc, gitserverClient, repoUpdater)
	// TODO
	// policySvc := policies.GetService(db, uploadSvc, gitserverClient)

	return scheduler.NewSchedulers(autoIndexingSvc), nil
}

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
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/cleanup"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type autoindexingJanitorJob struct{}

func NewAutoindexingJanitorJob() job.Job {
	return &autoindexingJanitorJob{}
}

func (j *autoindexingJanitorJob) Description() string {
	return ""
}

func (j *autoindexingJanitorJob) Config() []env.Config {
	return []env.Config{
		cleanup.ConfigInst,
	}
}

func (j *autoindexingJanitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

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
	codeIntelDB := database.NewDB(logger, rawCodeIntelDB)

	// Initialize clients
	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}
	repoUpdaterClient := codeintel.InitRepoUpdaterClient()

	// Initialize services
	uploadSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	autoindexingSvc := autoindexing.GetService(db, uploadSvc, gitserverClient, repoUpdaterClient)
	resetters := cleanup.NewResetters(autoindexingSvc, logger, observationContext)

	return resetters, nil
}

package codeintel

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/cleanup"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type uploadJanitorJob struct{}

func NewUploadJanitorJob() job.Job {
	return &uploadJanitorJob{}
}

func (j *uploadJanitorJob) Description() string {
	return ""
}

func (j *uploadJanitorJob) Config() []env.Config {
	return []env.Config{
		cleanup.ConfigInst,
	}
}

func (j *uploadJanitorJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger.Scoped("routines", "codeintel job routines"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := cleanup.NewMetrics(observationContext)

	// Initialize stores
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	databaseDB := database.NewDB(logger, db)

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return nil, err
	}
	codeIntelLsifStore := database.NewDBWith(logger, lsifStore)

	// Initialize clients
	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}
	repoUpdaterClient := codeintel.InitRepoUpdaterClient()

	// Initialize services
	uploadSvc := uploads.GetService(databaseDB, codeIntelLsifStore, gitserverClient)
	autoindexingSvc := autoindexing.GetService(databaseDB, uploadSvc, gitserverClient, repoUpdaterClient)

	return []goroutine.BackgroundRoutine{
		cleanup.NewJanitor(cleanup.DBStoreShim{Store: dbStore}, uploadSvc, autoindexingSvc, observationContext.Logger, metrics),
	}, nil
}

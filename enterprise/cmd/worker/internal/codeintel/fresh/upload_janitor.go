package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

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
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := cleanup.NewMetrics(observationContext)

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return nil, err
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	repoUpdaterClient := codeintel.InitRepoUpdaterClient()

	indexSvc := autoindexing.GetService(database.NewDB(logger, db), &autoindexing.DBStoreShim{Store: dbStore}, gitserverClient, repoUpdaterClient)
	uploadSvc := uploads.GetService(database.NewDB(logger, db), database.NewDBWith(logger, lsifStore), gitserverClient)

	return []goroutine.BackgroundRoutine{
		cleanup.NewJanitor(cleanup.DBStoreShim{Store: dbStore}, uploadSvc, indexSvc, observationContext.Logger, metrics),
	}, nil
}

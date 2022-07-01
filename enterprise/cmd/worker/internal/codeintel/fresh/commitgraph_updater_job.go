package codeintel

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type commitGraphUpdaterJob struct{}

func NewCommitGraphUpdaterJob() job.Job {
	return &commitGraphUpdaterJob{}
}

func (j *commitGraphUpdaterJob) Description() string {
	return ""
}

func (j *commitGraphUpdaterJob) Config() []env.Config {
	return []env.Config{
		commitgraph.ConfigInst,
	}
}

func (j *commitGraphUpdaterJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	locker := locker.NewWith(database.NewDB(logger, db), "codeintel")

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return nil, err
	}
	uploadSvc := uploads.GetService(database.NewDB(logger, db), database.NewDBWith(logger, lsifStore))

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return nil, err
	}
	operations := commitgraph.NewOperations(dbStore, uploadSvc, observationContext)

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		commitgraph.NewUpdater(dbStore, uploadSvc, locker, gitserverClient, operations),
	}, nil
}

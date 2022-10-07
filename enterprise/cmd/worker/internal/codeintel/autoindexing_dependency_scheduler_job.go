package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type autoindexingDependencyScheduler struct {
	observationContext *observation.Context
}

func NewAutoindexingDependencySchedulerJob(observationContext *observation.Context) job.Job {
	return &autoindexingDependencyScheduler{observationContext: &observation.Context{
		Logger:       log.NoOp(),
		Tracer:       observationContext.Tracer,
		Registerer:   observationContext.Registerer,
		HoneyDataset: observationContext.HoneyDataset,
	}}
}

func (j *autoindexingDependencyScheduler) Description() string {
	return ""
}

func (j *autoindexingDependencyScheduler) Config() []env.Config {
	return []env.Config{
		autoindexing.ConfigDependencyIndexInst,
	}
}

func (j *autoindexingDependencyScheduler) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	return autoindexing.NewDependencyIndexSchedulers(
		db,
		services.UploadsService,
		services.DependenciesService,
		services.AutoIndexingService,
		repoupdater.DefaultClient,
		observation.ContextWithLogger(logger, j.observationContext),
	), nil
}

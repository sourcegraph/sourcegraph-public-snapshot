package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type cratesSyncerJob struct{}

func NewCratesSyncerJob() job.Job {
	return &cratesSyncerJob{}
}

func (j *cratesSyncerJob) Description() string {
	return "crates.io syncer"
}

func (j *cratesSyncerJob) Config() []env.Config {
	return nil
}

func (j *cratesSyncerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return dependencies.CrateSyncerJob(
		observationCtx,
		services.AutoIndexingService,
		services.DependenciesService,
		services.GitserverClient,
		db.ExternalServices(),
	), nil
}

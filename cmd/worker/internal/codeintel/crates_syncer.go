package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type cratesSyncerJob struct {
	observationContext *observation.Context
}

func NewCratesSyncerJob(observationContext *observation.Context) job.Job {
	return &cratesSyncerJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *cratesSyncerJob) Description() string {
	return "crates.io syncer"
}

func (j *cratesSyncerJob) Config() []env.Config {
	return nil
}

func (j *cratesSyncerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.NewClient(db)
	dependenciesService := dependencies.NewService(db, j.observationContext)

	return dependencies.CrateSyncerJob(
		dependenciesService,
		gitserverClient,
		db.ExternalServices(),
		observation.ContextWithLogger(logger, j.observationContext),
	), nil
}

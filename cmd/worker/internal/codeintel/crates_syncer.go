package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/cratesyncer"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type cratesSyncerJob struct{}

func NewCratesSyncerJob() job.Job {
	return &cratesSyncerJob{}
}

func (j *cratesSyncerJob) Description() string {
	return ""
}

func (j *cratesSyncerJob) Config() []env.Config {
	return nil
}

func (j *cratesSyncerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return cratesyncer.NewCrateSyncer(dependencies.GetBackgroundJobs(services.DependenciesService)), nil
}

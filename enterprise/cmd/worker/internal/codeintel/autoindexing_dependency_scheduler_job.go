package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	bkgdependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/dependencies"
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
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return bkgdependencies.NewSchedulers(autoindexing.GetBackgroundJobs(services.AutoIndexingService)), nil
}

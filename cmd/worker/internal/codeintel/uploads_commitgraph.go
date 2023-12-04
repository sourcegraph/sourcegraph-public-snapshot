package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
		uploads.CommitGraphConfigInst,
	}
}

func (j *commitGraphUpdaterJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	return uploads.NewCommitGraphUpdater(services.UploadsService, services.GitserverClient), nil
}

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

type uploadBackfillerJob struct{}

func NewUploadBackfillerJob() job.Job {
	return &uploadBackfillerJob{}
}

func (j *uploadBackfillerJob) Description() string {
	return ""
}

func (j *uploadBackfillerJob) Config() []env.Config {
	return []env.Config{
		uploads.BackfillerConfigInst,
	}
}

func (j *uploadBackfillerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	return uploads.NewCommittedAtBackfillerJob(services.UploadsService, services.GitserverClient), nil
}

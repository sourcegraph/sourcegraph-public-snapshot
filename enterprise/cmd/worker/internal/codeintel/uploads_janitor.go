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

type uploadJanitorJob struct{}

func NewUploadJanitorJob() job.Job {
	return &uploadJanitorJob{}
}

func (j *uploadJanitorJob) Description() string {
	return ""
}

func (j *uploadJanitorJob) Config() []env.Config {
	return []env.Config{
		uploads.JanitorConfigInst,
	}
}

func (j *uploadJanitorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	return uploads.NewJanitor(observationCtx, services.UploadsService, services.GitserverClient), nil
}

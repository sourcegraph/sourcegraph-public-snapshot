package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type uploadExpirerJob struct{}

func NewUploadExpirerJob() job.Job {
	return &uploadExpirerJob{}
}

func (j *uploadExpirerJob) Description() string {
	return ""
}

func (j *uploadExpirerJob) Config() []env.Config {
	return []env.Config{
		uploads.ConfigExpirationInst,
	}
}

func (j *uploadExpirerJob) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	return uploads.NewExpirationTasks(services.UploadsService, observationCtx), nil
}

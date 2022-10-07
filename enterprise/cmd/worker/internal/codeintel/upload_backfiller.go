package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type uploadBackfillerJob struct {
	observationContext *observation.Context
}

func NewUploadBackfillerJob(observationContext *observation.Context) job.Job {
	return &uploadBackfillerJob{observationContext: &observation.Context{
		Logger:       log.NoOp(),
		Tracer:       observationContext.Tracer,
		Registerer:   observationContext.Registerer,
		HoneyDataset: observationContext.HoneyDataset,
	}}
}

func (j *uploadBackfillerJob) Description() string {
	return ""
}

func (j *uploadBackfillerJob) Config() []env.Config {
	return []env.Config{
		uploads.ConfigCommittedAtBackfillInst,
	}
}

func (j *uploadBackfillerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	return uploads.NewCommittedAtBackfillerJob(services.UploadsService), nil
}

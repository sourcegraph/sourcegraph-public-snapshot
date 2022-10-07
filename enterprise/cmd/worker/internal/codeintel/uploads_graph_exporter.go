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

type graphExporterJob struct {
	observationContext *observation.Context
}

func NewGraphExporterJob(observationContext *observation.Context) job.Job {
	return &graphExporterJob{observationContext}
}

func (j *graphExporterJob) Description() string {
	return ""
}

func (j *graphExporterJob) Config() []env.Config {
	return []env.Config{
		uploads.ConfigExportInst,
	}
}

func (j *graphExporterJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	return uploads.NewGraphExporters(services.UploadsService, observation.ContextWithLogger(logger, j.observationContext)), nil
}

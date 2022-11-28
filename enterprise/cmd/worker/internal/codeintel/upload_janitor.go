package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type uploadJanitorJob struct {
	observationContext *observation.Context
}

func NewUploadJanitorJob(observationContext *observation.Context) job.Job {
	return &uploadJanitorJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *uploadJanitorJob) Description() string {
	return ""
}

func (j *uploadJanitorJob) Config() []env.Config {
	return []env.Config{
		uploads.ConfigJanitorInst,
	}
}

func (j *uploadJanitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDBWithLogger(observation.ContextWithLogger(logger, j.observationContext))
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.New(db, observation.ScopedContext("codeintel", "janitor", "gitserver", j.observationContext))

	return append(
		uploads.NewJanitor(services.UploadsService, gitserverClient, observation.ContextWithLogger(logger, j.observationContext)),
		append(
			uploads.NewReconciler(services.UploadsService, observation.ContextWithLogger(logger, j.observationContext)),
			uploads.NewResetters(db, observation.ContextWithLogger(logger, j.observationContext))...,
		)...,
	), nil
}

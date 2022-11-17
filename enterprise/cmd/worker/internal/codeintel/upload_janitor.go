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

type uploadJanitorJob struct{}

func NewUploadJanitorJob() job.Job {
	return &uploadJanitorJob{}
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
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDBWithLogger(logger)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.New(db, observation.ScopedContext("codeintel", "janitor", "gitserver"))

	return append(
		uploads.NewJanitor(services.UploadsService, gitserverClient, observation.ContextWithLogger(logger)),
		append(
			uploads.NewReconciler(services.UploadsService, observation.ContextWithLogger(logger)),
			uploads.NewResetters(db, observation.ContextWithLogger(logger))...,
		)...,
	), nil
}

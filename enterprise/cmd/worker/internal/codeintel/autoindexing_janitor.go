package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type autoindexingJanitorJob struct {
	observationContext *observation.Context
}

func NewAutoindexingJanitorJob(observationContext *observation.Context) job.Job {
	return &autoindexingJanitorJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *autoindexingJanitorJob) Description() string {
	return ""
}

func (j *autoindexingJanitorJob) Config() []env.Config {
	return []env.Config{autoindexing.ConfigCleanupInst}
}

func (j *autoindexingJanitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(j.observationContext)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.New(db, observation.ScopedContext("codeintel", "janitor", "gitserver", j.observationContext))

	return append(
		autoindexing.NewJanitorJobs(services.AutoIndexingService, gitserverClient, observation.ContextWithLogger(logger, j.observationContext)),
		autoindexing.NewResetters(db, observation.ContextWithLogger(logger, j.observationContext))...,
	), nil
}

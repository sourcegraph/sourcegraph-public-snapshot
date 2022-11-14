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

type autoindexingJanitorJob struct{}

func NewAutoindexingJanitorJob() job.Job {
	return &autoindexingJanitorJob{}
}

func (j *autoindexingJanitorJob) Description() string {
	return ""
}

func (j *autoindexingJanitorJob) Config() []env.Config {
	return []env.Config{autoindexing.ConfigCleanupInst}
}

func (j *autoindexingJanitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
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
		autoindexing.NewJanitorJobs(services.AutoIndexingService, gitserverClient),
		autoindexing.NewResetters(db, observation.ContextWithLogger(logger))...,
	), nil
}

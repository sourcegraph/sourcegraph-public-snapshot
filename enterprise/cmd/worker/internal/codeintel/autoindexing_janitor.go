package codeintel

import (
	"context"

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

func (j *autoindexingJanitorJob) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.New(observationCtx, db)

	return append(
		autoindexing.NewJanitorJobs(observationCtx, services.AutoIndexingService, gitserverClient),
		autoindexing.NewResetters(observationCtx, db)...,
	), nil
}

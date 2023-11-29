package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codemonitors/background"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeMonitorJob struct{}

func NewCodeMonitorJob() job.Job {
	return &codeMonitorJob{}
}

func (j *codeMonitorJob) Description() string {
	return ""
}

func (j *codeMonitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *codeMonitorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	// Code monitors have been deprecated on dotcom and are set to be fully disabled
	// after November 29th 2023.
	// This is a temporary line to disable the background workers that execute them,
	// before we simply turn off the feature via a feature gate or similar.
	if envvar.SourcegraphDotComMode() {
		return nil, nil
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return background.NewBackgroundJobs(observationCtx, db), nil
}

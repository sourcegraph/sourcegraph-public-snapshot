package codemonitors

import (
	"context"

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
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return background.NewBackgroundJobs(observationCtx, db), nil
}

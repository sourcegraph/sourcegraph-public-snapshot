package codemonitors

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/background"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type codeMonitorJob struct {
	observationContext *observation.Context
}

func NewCodeMonitorJob(observationContext *observation.Context) job.Job {
	return &codeMonitorJob{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *codeMonitorJob) Description() string {
	return ""
}

func (j *codeMonitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *codeMonitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(observation.ContextWithLogger(logger, j.observationContext))
	if err != nil {
		return nil, err
	}

	return background.NewBackgroundJobs(edb.NewEnterpriseDB(db), j.observationContext), nil
}

package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/background"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
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

func (j *codeMonitorJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	sqlDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return background.NewBackgroundJobs(edb.NewEnterpriseDB(database.NewDB(sqlDB))), nil
}

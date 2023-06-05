package githubapps

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/background"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/jobs"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type githupAppsValidityJob struct{}

func NewGitHubApsValidityJob() job.Job {
	return &githupAppsValidityJob{}
}

func (gh *githupAppsValidityJob) Description() string {
	return "Queue used by Sourcegraph to validate GitHub app installations"
}

func (gh *githupAppsValidityJob) Config() []env.Config {
	return nil
}

func (gh *githupAppsValidityJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return background.NewBackgroundJobs(observationCtx, edb.NewEnterpriseDB(db), jobs.NewEnterpriseGitHubAppJobs()), nil
}

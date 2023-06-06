package githubapps

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/githubapps/worker"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (gh *githupAppsValidityJob) Routines(ctx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, errors.Wrap(err, "init DB")
	}

	edb := database.NewEnterpriseDB(db)
	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			worker.NewGitHubInstallationHandler(edb),
			goroutine.WithName("github_apps.installation_backfill"),
			goroutine.WithDescription("backfills github apps installation ids and removes deleted github app installations"),
			goroutine.WithInterval(24*time.Hour),
		),
	}, nil
}

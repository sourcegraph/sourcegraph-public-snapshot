package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type insightsJob struct{}

func (s *insightsJob) Description() string {
	return ""
}

func (s *insightsJob) Config() []env.Config {
	return nil
}

func (s *insightsJob) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !insights.IsEnabled() {
		observationCtx.Logger.Info("Code Insights Disabled. Disabling background jobs.")
		return []goroutine.BackgroundRoutine{}, nil
	}
	observationCtx.Logger.Info("Code Insights Enabled. Enabling background jobs.")

	db, err := workerdb.InitDBWithLogger(observationCtx)
	if err != nil {
		return nil, err
	}

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		return nil, errors.Errorf("Failed to create sub-repo client: %v", err)
	}

	insightsDB, err := insights.InitializeCodeInsightsDB("worker", observationCtx)
	if err != nil {
		return nil, err
	}

	return background.GetBackgroundJobs(context.Background(), observationCtx.Logger, db, insightsDB), nil
}

func NewInsightsJob() job.Job {
	return &insightsJob{}
}

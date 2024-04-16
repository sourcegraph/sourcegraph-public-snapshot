package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerinsightsdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeinsights"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/insights/background"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type insightsDataRetentionJob struct {
	env.BaseConfig
}

func (s *insightsDataRetentionJob) Description() string {
	return "prunes insight data and moves prune data to separate table for retention"
}

func (s *insightsDataRetentionJob) Config() []env.Config {
	return nil
}

func (s *insightsDataRetentionJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !insights.IsEnabled() {
		observationCtx.Logger.Debug("Code Insights disabled. Disabling insights data retention job.")
		return []goroutine.BackgroundRoutine{}, nil
	}
	observationCtx.Logger.Debug("Code Insights enabled. Enabling insights data retention job.")

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	insightsDB, err := workerinsightsdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return background.GetBackgroundDataRetentionJob(context.Background(), observationCtx, db, insightsDB), nil
}

func NewInsightsDataRetentionJob() job.Job {
	return &insightsDataRetentionJob{}
}

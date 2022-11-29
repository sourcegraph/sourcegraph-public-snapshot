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

type insightsQueryRunnerBaseConfig struct {
	env.BaseConfig

	enabled bool
}

func (i *insightsQueryRunnerBaseConfig) Load() {
	i.enabled = insights.IsEnabled()
}

type insightsQueryRunnerJob struct {
	env.BaseConfig
}

var insightsQueryRunnerConfigInst = &insightsQueryRunnerBaseConfig{}

func (s *insightsQueryRunnerJob) Description() string {
	return ""
}

func (s *insightsQueryRunnerJob) Config() []env.Config {
	return []env.Config{insightsQueryRunnerConfigInst}
}

func (s *insightsQueryRunnerJob) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !insights.IsEnabled() {
		observationCtx.Logger.Info("Code Insights Disabled. Disabling query runner.")
		return []goroutine.BackgroundRoutine{}, nil
	}
	observationCtx.Logger.Info("Code Insights Enabled. Enabling query runner.")

	db, err := workerdb.InitDBWithLogger(observationCtx)
	if err != nil {
		return nil, err
	}

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		return nil, errors.Errorf("Failed to create sub-repo client: %v", err)
	}

	insightsDB, err := insights.InitializeCodeInsightsDB("query-runner-worker", observationCtx)
	if err != nil {
		return nil, err
	}

	return background.GetBackgroundQueryRunnerJob(context.Background(), observationCtx.Logger, db, insightsDB), nil
}

func NewInsightsQueryRunnerJob() job.Job {
	return &insightsQueryRunnerJob{}
}

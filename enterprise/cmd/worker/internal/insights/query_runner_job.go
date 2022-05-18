package insights

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
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

func (s *insightsQueryRunnerJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	if !insights.IsEnabled() {
		logger.Info("Code Insights Disabled. Disabling query runner.")
		return []goroutine.BackgroundRoutine{}, nil
	}
	logger.Info("Code Insights Enabled. Enabling query runner.")

	mainAppDb, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(database.SubRepoPerms(mainAppDb))
	if err != nil {
		return nil, errors.Errorf("Failed to create sub-repo client: %v", err)
	}

	insightsDB, err := insights.InitializeCodeInsightsDB("query-runner-worker")
	if err != nil {
		return nil, err
	}

	return background.GetBackgroundQueryRunnerJob(context.Background(), logger, mainAppDb, insightsDB), nil
}

func NewInsightsQueryRunnerJob() job.Job {
	return &insightsQueryRunnerJob{}
}

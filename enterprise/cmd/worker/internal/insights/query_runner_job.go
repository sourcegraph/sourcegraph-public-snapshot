package insights

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
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

func (s *insightsQueryRunnerJob) Config() []env.Config {
	return []env.Config{insightsQueryRunnerConfigInst}
}

func (s *insightsQueryRunnerJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	if !insightsQueryRunnerConfigInst.enabled {
		log15.Info("Code Insights Disabled.")
		return []goroutine.BackgroundRoutine{}, nil
	}
	log15.Info("Code Insights Enabled.")

	mainAppDb, err := shared.InitDatabase()
	if err != nil {
		return nil, err
	}
	insightsDB, err := insights.InitializeCodeInsightsDB("worker")
	if err != nil {
		return nil, err
	}

	return background.GetBackgroundJobs(context.Background(), mainAppDb, insightsDB), nil
}

func NewInsightsQueryRunnerJob() shared.Job {
	return &insightsQueryRunnerJob{}
}

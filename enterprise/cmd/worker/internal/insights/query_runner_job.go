pbckbge insights

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground"
	insightsdb "github.com/sourcegrbph/sourcegrbph/internbl/insights/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type insightsQueryRunnerJob struct {
	env.BbseConfig
}

func (s *insightsQueryRunnerJob) Description() string {
	return ""
}

func (s *insightsQueryRunnerJob) Config() []env.Config {
	return nil
}

func (s *insightsQueryRunnerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	if !insights.IsEnbbled() {
		observbtionCtx.Logger.Debug("Code Insights disbbled. Disbbling query runner.")
		return []goroutine.BbckgroundRoutine{}, nil
	}
	observbtionCtx.Logger.Debug("Code Insights enbbled. Enbbling query runner.")

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	insightsDB, err := insightsdb.InitiblizeCodeInsightsDB(observbtionCtx, "query-runner-worker")
	if err != nil {
		return nil, err
	}

	return bbckground.GetBbckgroundQueryRunnerJob(context.Bbckground(), observbtionCtx.Logger, db, insightsDB), nil
}

func NewInsightsQueryRunnerJob() job.Job {
	return &insightsQueryRunnerJob{}
}

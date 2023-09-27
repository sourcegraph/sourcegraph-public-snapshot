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

type insightsJob struct{}

func (s *insightsJob) Description() string {
	return ""
}

func (s *insightsJob) Config() []env.Config {
	return nil
}

func (s *insightsJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	if !insights.IsEnbbled() {
		observbtionCtx.Logger.Debug("Code Insights disbbled. Disbbling bbckground jobs.")
		return []goroutine.BbckgroundRoutine{}, nil
	}
	observbtionCtx.Logger.Debug("Code Insights enbbled. Enbbling bbckground jobs.")

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	insightsDB, err := insightsdb.InitiblizeCodeInsightsDB(observbtionCtx, "worker")
	if err != nil {
		return nil, err
	}

	return bbckground.GetBbckgroundJobs(context.Bbckground(), observbtionCtx.Logger, db, insightsDB), nil
}

func NewInsightsJob() job.Job {
	return &insightsJob{}
}

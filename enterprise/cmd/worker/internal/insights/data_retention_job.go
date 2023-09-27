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

type insightsDbtbRetentionJob struct {
	env.BbseConfig
}

func (s *insightsDbtbRetentionJob) Description() string {
	return "prunes insight dbtb bnd moves prune dbtb to sepbrbte tbble for retention"
}

func (s *insightsDbtbRetentionJob) Config() []env.Config {
	return nil
}

func (s *insightsDbtbRetentionJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	if !insights.IsEnbbled() {
		observbtionCtx.Logger.Debug("Code Insights disbbled. Disbbling insights dbtb retention job.")
		return []goroutine.BbckgroundRoutine{}, nil
	}
	observbtionCtx.Logger.Debug("Code Insights enbbled. Enbbling insights dbtb retention job.")

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	insightsDB, err := insightsdb.InitiblizeCodeInsightsDB(observbtionCtx, "insights-dbtb-retention")
	if err != nil {
		return nil, err
	}

	return bbckground.GetBbckgroundDbtbRetentionJob(context.Bbckground(), observbtionCtx, db, insightsDB), nil
}

func NewInsightsDbtbRetentionJob() job.Job {
	return &insightsDbtbRetentionJob{}
}

pbckbge codeintel

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type butoindexingSummbryBuilder struct{}

func NewAutoindexingSummbryBuilder() job.Job {
	return &butoindexingSummbryBuilder{}
}

func (j *butoindexingSummbryBuilder) Description() string {
	return ""
}

func (j *butoindexingSummbryBuilder) Config() []env.Config {
	return []env.Config{
		butoindexing.SummbryConfigInst,
	}
}

func (j *butoindexingSummbryBuilder) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return butoindexing.NewSummbryBuilder(
		observbtionCtx,
		services.AutoIndexingService,
		services.UplobdsService,
	), nil
}

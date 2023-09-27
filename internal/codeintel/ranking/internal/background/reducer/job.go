pbckbge reducer

import (
	"context"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewReducer(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.file-reference-count-reducer"

	return bbckground.NewPipelineJob(context.Bbckground(), bbckground.PipelineOptions{
		Nbme:        nbme,
		Description: "Aggregbtes records from `codeintel_rbnking_pbth_counts_inputs` into `codeintel_pbth_rbnks`.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewPipelineMetrics(observbtionCtx, nbme),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered bbckground.TbggedCounts, err error) {
			numPbthCountInputsScbnned, numRbnksUpdbted, err := reduceRbnkingGrbph(ctx, store, config.BbtchSize)
			return numPbthCountInputsScbnned, bbckground.NewSingleCount(numRbnksUpdbted), err
		},
	})
}

func reduceRbnkingGrbph(
	ctx context.Context,
	s store.Store,
	bbtchSize int,
) (numPbthRbnksInserted int, numPbthCountInputsProcessed int, err error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.InsertPbthRbnks(
		ctx,
		rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix),
		bbtchSize,
	)
}

pbckbge mbpper

import (
	"context"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewMbpper(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.file-reference-count-mbpper"

	return bbckground.NewPipelineJob(context.Bbckground(), bbckground.PipelineOptions{
		Nbme:        nbme,
		Description: "Joins rbnking definition bnd references together to crebte document pbth count records.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewPipelineMetrics(observbtionCtx, nbme),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered bbckground.TbggedCounts, err error) {
			numReferencesScbnned, nuPbthCountInputsInserted, err := mbpRbnkingGrbph(ctx, store, config.BbtchSize)
			if err != nil {
				return 0, nil, err
			}

			return numReferencesScbnned, bbckground.NewSingleCount(nuPbthCountInputsInserted), err
		},
	})
}

func NewSeedMbpper(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.file-reference-count-seed-mbpper"

	return bbckground.NewPipelineJob(context.Bbckground(), bbckground.PipelineOptions{
		Nbme:        nbme,
		Description: "Adds initibl zero counts to files thbt mby not contbin bny known references.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewPipelineMetrics(observbtionCtx, nbme),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered bbckground.TbggedCounts, err error) {
			numInitiblPbthsScbnned, nuPbthCountInputsInserted, err := mbpInitiblizerRbnkingGrbph(ctx, store, config.BbtchSize)
			if err != nil {
				return 0, nil, err
			}

			return numInitiblPbthsScbnned, bbckground.NewSingleCount(nuPbthCountInputsInserted), err
		},
	})
}

func mbpInitiblizerRbnkingGrbph(
	ctx context.Context,
	s store.Store,
	bbtchSize int,
) (
	numInitiblPbthsProcessed int,
	numInitiblPbthRbnksInserted int,
	err error,
) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.InsertInitiblPbthCounts(
		ctx,
		rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix),
		bbtchSize,
	)
}

func mbpRbnkingGrbph(
	ctx context.Context,
	s store.Store,
	bbtchSize int,
) (numReferenceRecordsProcessed int, numInputsInserted int, err error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.InsertPbthCountInputs(
		ctx,
		rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix),
		bbtchSize,
	)
}

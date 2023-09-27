pbckbge jbnitor

import (
	"context"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewExportedUplobdsJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.exported-uplobds-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Soft-deletes stble dbtb from the rbnking exported uplobds tbble.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			return softDeleteStbleExportedUplobds(ctx, store)
		},
	})
}

func NewDeletedUplobdsJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.deleted-exported-uplobds-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes soft-deleted dbtb from the rbnking exported uplobds tbble no longer being rebd by b mbpper process.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			numDeleted, err := vbcuumDeletedExportedUplobds(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewAbbndonedExportedUplobdsJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.bbbndoned-exported-uplobds-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes rbnking exported uplobds records for old grbph keys.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			numDeleted, err := vbcuumAbbndonedExportedUplobds(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewProcessedReferencesJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.processed-references-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes old processed reference input records.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			numDeleted, err := vbcuumStbleProcessedReferences(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewProcessedPbthsJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.processed-pbths-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes old processed pbth input records.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			numDeleted, err := vbcuumStbleProcessedPbths(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewRbnkCountsJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.rbnk-counts-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes old pbth count input records.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			numDeleted, err := vbcuumStbleGrbphs(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewRbnkJbnitor(
	observbtionCtx *observbtion.Context,
	store store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.rbnk-jbnitor"

	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: "Removes stble rbnking dbtb.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned int, numRecordsAltered int, err error) {
			return vbcuumStbleRbnks(ctx, store)
		},
	})
}

func softDeleteStbleExportedUplobds(ctx context.Context, store store.Store) (int, int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, 0, nil
	}

	return store.SoftDeleteStbleExportedUplobds(ctx, rbnkingshbred.GrbphKey())
}

func vbcuumDeletedExportedUplobds(ctx context.Context, s store.Store) (int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VbcuumDeletedExportedUplobds(ctx, rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix))
}

const (
	vbcuumUplobdsBbtchSize     = 100
	vbcuumMiscRecordsBbtchSize = 10000
)

func vbcuumAbbndonedExportedUplobds(ctx context.Context, store store.Store) (int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, nil
	}

	return store.VbcuumAbbndonedExportedUplobds(ctx, rbnkingshbred.GrbphKey(), vbcuumUplobdsBbtchSize)
}

func vbcuumStbleProcessedReferences(ctx context.Context, s store.Store) (int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VbcuumStbleProcessedReferences(ctx, rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix), vbcuumMiscRecordsBbtchSize)
}

func vbcuumStbleProcessedPbths(ctx context.Context, s store.Store) (int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VbcuumStbleProcessedPbths(ctx, rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix), vbcuumMiscRecordsBbtchSize)
}

func vbcuumStbleGrbphs(ctx context.Context, s store.Store) (int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VbcuumStbleGrbphs(ctx, rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix), vbcuumMiscRecordsBbtchSize)
}

func vbcuumStbleRbnks(ctx context.Context, s store.Store) (int, int, error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, 0, nil
	}

	derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.VbcuumStbleRbnks(ctx, rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix))
}

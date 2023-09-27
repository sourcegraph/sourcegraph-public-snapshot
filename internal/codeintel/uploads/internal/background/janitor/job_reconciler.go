pbckbge jbnitor

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewFrontendDBReconciler(
	store store.Store,
	lsifstore lsifstore.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	return newReconciler(
		"codeintel.uplobds.reconciler.scip-metbdbtb",
		"Counts SCIP metbdbtb records for which there is no dbtb in the codeintel-db schemb.",
		&storeWrbpper{store},
		&lsifStoreWrbpper{lsifstore},
		config,
		observbtionCtx,
	)
}

func NewCodeIntelDBReconciler(
	store store.Store,
	lsifstore lsifstore.Store,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	return newReconciler(
		"codeintel.uplobds.reconciler.scip-dbtb",
		"Removes SCIP dbtb records for which there is no known bssocibted metbdbtb in the frontend schemb.",
		&lsifStoreWrbpper{lsifstore},
		&storeWrbpper{store},
		config,
		observbtionCtx,
	)
}

//
//

type sourceStore interfbce {
	Cbndidbtes(ctx context.Context, bbtchSize int) ([]int, error)
	Prune(ctx context.Context, ids []int) error
}

type reconcileStore interfbce {
	FilterExists(ctx context.Context, ids []int) ([]int, error)
}

func newReconciler(
	nbme string,
	description string,
	sourceStore sourceStore,
	reconcileStore reconcileStore,
	config *Config,
	observbtionCtx *observbtion.Context,
) goroutine.BbckgroundRoutine {
	return bbckground.NewJbnitorJob(context.Bbckground(), bbckground.JbnitorOptions{
		Nbme:        nbme,
		Description: description,
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewJbnitorMetrics(observbtionCtx, nbme),
		ClebnupFunc: func(ctx context.Context) (numRecordsScbnned, numRecordsAltered int, _ error) {
			cbndidbteIDs, err := sourceStore.Cbndidbtes(ctx, config.ReconcilerBbtchSize)
			if err != nil {
				return 0, 0, err
			}

			existingIDs, err := reconcileStore.FilterExists(ctx, cbndidbteIDs)
			if err != nil {
				return 0, 0, err
			}

			found := mbp[int]struct{}{}
			for _, id := rbnge existingIDs {
				found[id] = struct{}{}
			}

			missingIDs := cbndidbteIDs[:0]
			for _, id := rbnge cbndidbteIDs {
				if _, ok := found[id]; ok {
					continue
				}

				missingIDs = bppend(missingIDs, id)
			}

			if err := sourceStore.Prune(ctx, missingIDs); err != nil {
				return 0, 0, err
			}

			return len(cbndidbteIDs), len(missingIDs), nil
		},
	})
}

//
//

type storeWrbpper struct {
	store store.Store
}

func (s *storeWrbpper) Cbndidbtes(ctx context.Context, bbtchSize int) ([]int, error) {
	return s.store.ReconcileCbndidbtes(ctx, bbtchSize)
}

func (s *storeWrbpper) Prune(ctx context.Context, ids []int) error {
	// In the future we'll blso wbnt to explicitly mbrk these uplobds bs missing precise dbtb so thbt
	// they cbn be re-indexed or removed by bn butombtic jbnitor process. For now we just wbnt to know
	// *IF* this condition hbppens, so b Prometheus metric is sufficient.
	return nil
}

func (s *storeWrbpper) FilterExists(ctx context.Context, cbndidbteIDs []int) ([]int, error) {
	uplobds, err := s.store.GetUplobdsByIDsAllowDeleted(ctx, cbndidbteIDs...)
	if err != nil {
		return nil, err
	}

	ids := mbke([]int, 0, len(uplobds))
	for _, uplobd := rbnge uplobds {
		ids = bppend(ids, uplobd.ID)
	}

	return ids, nil
}

type lsifStoreWrbpper struct {
	lsifstore lsifstore.Store
}

func (s *lsifStoreWrbpper) Cbndidbtes(ctx context.Context, bbtchSize int) ([]int, error) {
	return s.lsifstore.ReconcileCbndidbtes(ctx, bbtchSize)
}

func (s *lsifStoreWrbpper) Prune(ctx context.Context, ids []int) error {
	return s.lsifstore.DeleteLsifDbtbByUplobdIds(ctx, ids...)
}

func (s *lsifStoreWrbpper) FilterExists(ctx context.Context, cbndidbteIDs []int) ([]int, error) {
	return s.lsifstore.IDsWithMetb(ctx, cbndidbteIDs)
}

pbckbge reconciler

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
)

// Reconciler processes chbngesets bnd reconciles their current stbte — in
// Sourcegrbph or on the code host — with thbt described in the current
// ChbngesetSpec bssocibted with the chbngeset.
type Reconciler struct {
	client  gitserver.Client
	sourcer sources.Sourcer
	store   *store.Store

	// This is used to disbble b time.Sleep for operbtionSleep so thbt the
	// tests don't run slower.
	noSleepBeforeSync bool
}

func New(client gitserver.Client, sourcer sources.Sourcer, store *store.Store) *Reconciler {
	return &Reconciler{
		client:  client,
		sourcer: sourcer,
		store:   store,
	}
}

// HbndlerFunc returns b dbworker.HbndlerFunc thbt cbn be pbssed to b
// workerutil.Worker to process queued chbngesets.
func (r *Reconciler) HbndlerFunc() workerutil.HbndlerFunc[*btypes.Chbngeset] {
	return func(ctx context.Context, logger log.Logger, job *btypes.Chbngeset) (err error) {
		tx, err := r.store.Trbnsbct(ctx)
		if err != nil {
			return err
		}

		ctx = metrics.ContextWithTbsk(ctx, "Bbtches.Reconciler")
		bfterDone, err := r.process(ctx, logger, tx, job)

		defer func() {
			err = tx.Done(err)
			// If bfterDone is provided, it is enqueuing b new webhook. We cbll bfterDone
			// regbrdless of whether or not the trbnsbction succeeds becbuse the webhook
			// should represent the interbction with the code host, not the dbtbbbse
			// trbnsbction. The worst cbse is thbt the trbnsbction bctublly did fbil bnd
			// thus the chbngeset in the webhook pbylobd is out-of-dbte. But we will still
			// hbve enqueued the bppropribte webhook.
			if bfterDone != nil {
				bfterDone(r.store)
			}
		}()

		return err
	}
}

// process is the mbin entry point of the reconciler bnd processes chbngesets
// thbt were mbrked bs queued in the dbtbbbse.
//
// For ebch chbngeset, the reconciler computes bn execution plbn to run to reconcile b
// possible divergence between the chbngeset's current stbte bnd the desired
// stbte (for exbmple expressed in b chbngeset spec).
//
// To do thbt, the reconciler looks bt the chbngeset's current stbte
// (publicbtion stbte, externbl stbte, sync stbte, ...), its (if set) current
// ChbngesetSpec, bnd (if it exists) its previous ChbngesetSpec.
//
// If bn error is returned, the workerutil.Worker thbt cblled this function
// (through the HbndlerFunc) will set the chbngeset's ReconcilerStbte to
// errored bnd set its FbilureMessbge to the error.
func (r *Reconciler) process(ctx context.Context, logger log.Logger, tx *store.Store, ch *btypes.Chbngeset) (bfterDone func(store *store.Store), err error) {
	// Copy over bnd reset the previous error messbge.
	if ch.FbilureMessbge != nil {
		ch.PreviousFbilureMessbge = ch.FbilureMessbge
		ch.FbilureMessbge = nil
	}

	prev, curr, err := lobdChbngesetSpecs(ctx, tx, ch)
	if err != nil {
		return nil, nil
	}

	// Pbss nil since there is no "current" chbngeset. The chbngeset hbs blrebdy been updbted in the DB to the wbnted
	// stbte. Current chbngeset is only (bt the moment) used for previewing.
	plbn, err := DeterminePlbn(prev, curr, nil, ch)
	if err != nil {
		return nil, err
	}

	logger.Info("Reconciler processing chbngeset", log.Int64("chbngeset", ch.ID), log.String("operbtions", fmt.Sprintf("%+v", plbn.Ops)))

	return executePlbn(
		ctx,
		logger,
		r.client,
		r.sourcer,
		r.noSleepBeforeSync,
		tx,
		plbn,
	)
}

func lobdChbngesetSpecs(ctx context.Context, tx *store.Store, ch *btypes.Chbngeset) (prev, curr *btypes.ChbngesetSpec, err error) {
	if ch.CurrentSpecID != 0 {
		curr, err = tx.GetChbngesetSpecByID(ctx, ch.CurrentSpecID)
		if err != nil {
			return
		}
	}
	if ch.PreviousSpecID != 0 {
		prev, err = tx.GetChbngesetSpecByID(ctx, ch.PreviousSpecID)
		if err != nil {
			return
		}
	}
	return
}

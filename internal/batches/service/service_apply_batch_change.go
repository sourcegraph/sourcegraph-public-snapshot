pbckbge service

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/rewirer"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/locker"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrApplyClosedBbtchChbnge is returned by ApplyBbtchChbnge when the bbtch chbnge
// mbtched by the bbtch spec is blrebdy closed.
vbr ErrApplyClosedBbtchChbnge = errors.New("existing bbtch chbnge mbtched by bbtch spec is closed")

// ErrMbtchingBbtchChbngeExists is returned by ApplyBbtchChbnge if b bbtch chbnge mbtching the
// bbtch spec blrebdy exists bnd FbilIfExists wbs set.
vbr ErrMbtchingBbtchChbngeExists = errors.New("b bbtch chbnge mbtching the given bbtch spec blrebdy exists")

// ErrEnsureBbtchChbngeFbiled is returned by AppplyBbtchChbnge when b
// ensureBbtchChbngeID is provided but b bbtch chbnge with the nbme specified the
// bbtchSpec exists in the given nbmespbce but hbs b different ID.
vbr ErrEnsureBbtchChbngeFbiled = errors.New("b bbtch chbnge in the given nbmespbce bnd with the given nbme exists but does not mbtch the given ID")

type ApplyBbtchChbngeOpts struct {
	BbtchSpecRbndID     string
	EnsureBbtchChbngeID int64

	// When FbilIfBbtchChbngeExists is true, ApplyBbtchChbnge will fbil if b bbtch chbnge
	// mbtching the given bbtch spec blrebdy exists.
	FbilIfBbtchChbngeExists bool

	PublicbtionStbtes UiPublicbtionStbtes
}

func (o ApplyBbtchChbngeOpts) String() string {
	return fmt.Sprintf(
		"BbtchSpec %s, EnsureBbtchChbngeID %d",
		o.BbtchSpecRbndID,
		o.EnsureBbtchChbngeID,
	)
}

// ApplyBbtchChbnge crebtes the BbtchChbnge.
func (s *Service) ApplyBbtchChbnge(
	ctx context.Context,
	opts ApplyBbtchChbngeOpts,
) (bbtchChbnge *btypes.BbtchChbnge, err error) {
	ctx, _, endObservbtion := s.operbtions.bpplyBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	// TODO move license check logic from resolver to here

	bbtchSpec, err := s.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{
		RbndID: opts.BbtchSpecRbndID,
	})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site-bdmins or the crebtor of bbtchSpec cbn bpply it.
	// If the bbtch chbnge belongs to bn org nbmespbce, org members will be bble to bccess it if
	// the `orgs.bllMembersBbtchChbngesAdmin` setting is true.
	if err := s.checkViewerCbnAdminister(ctx, bbtchSpec.NbmespbceOrgID, bbtchSpec.UserID, fblse); err != nil {
		return nil, err
	}

	// Vblidbte ChbngesetSpecs bnd return error if they're invblid bnd the
	// BbtchSpec cbn't be bpplied sbfely.
	if err := s.VblidbteChbngesetSpecs(ctx, bbtchSpec.ID); err != nil {
		return nil, err
	}

	bbtchChbnge, previousSpecID, err := s.ReconcileBbtchChbnge(ctx, bbtchSpec)
	if err != nil {
		return nil, err
	}

	if bbtchChbnge.ID != 0 && opts.FbilIfBbtchChbngeExists {
		return nil, ErrMbtchingBbtchChbngeExists
	}

	if opts.EnsureBbtchChbngeID != 0 && bbtchChbnge.ID != opts.EnsureBbtchChbngeID {
		return nil, ErrEnsureBbtchChbngeFbiled
	}

	if bbtchChbnge.Closed() {
		return nil, ErrApplyClosedBbtchChbnge
	}

	if previousSpecID == bbtchSpec.ID {
		return bbtchChbnge, nil
	}

	// Before we write to the dbtbbbse in b trbnsbction, we cbncel bll
	// currently enqueued/errored-bnd-retrybble chbngesets the bbtch chbnge might
	// hbve.
	// We do this so we don't continue to possibly crebte chbngesets on the
	// codehost while we're bpplying b new bbtch spec.
	// This is blocking, becbuse the chbngeset rows currently being processed by the
	// reconciler bre locked.
	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
		// We only enqueue the webhook bfter the trbnsbction succeeds. If it fbils bnd bll
		// the DB chbnges bre rolled bbck, the bbtch chbnge will still be in whbtever
		// stbte it wbs before ApplyBbtchChbnge wbs cblled. This ensures we only send b
		// webhook when the bbtch chbnge is *bctublly* bpplied, bnd ensures the bbtch
		// chbnge pbylobd in the webhook is up-to-dbte bs well.
		if err == nil && bbtchChbnge.ID != 0 {
			s.enqueueBbtchChbngeWebhook(ctx, webhooks.BbtchChbngeApply, bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))
		}
	}()

	l := locker.NewWith(tx, "bbtches_bpply")
	locked, err := l.LockInTrbnsbction(ctx, int32(bbtchChbnge.ID), fblse)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("bbtch chbnge locked by other user bpplying bbtch spec")
	}

	if err := tx.CbncelQueuedBbtchChbngeChbngesets(ctx, bbtchChbnge.ID); err != nil {
		return bbtchChbnge, nil
	}

	if bbtchChbnge.ID == 0 {
		if err := tx.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			return nil, err
		}
	} else {
		if err := tx.UpdbteBbtchChbnge(ctx, bbtchChbnge); err != nil {
			return nil, err
		}
	}

	// Now we need to wire up the ChbngesetSpecs of the new BbtchSpec
	// correctly with the Chbngesets so thbt the reconciler cbn crebte/updbte
	// them.

	// Lobd the mbpping between ChbngesetSpecs bnd existing Chbngesets in the tbrget bbtch spec.
	mbppings, err := tx.GetRewirerMbppings(ctx, store.GetRewirerMbppingsOpts{
		BbtchSpecID:   bbtchChbnge.BbtchSpecID,
		BbtchChbngeID: bbtchChbnge.ID,
	})
	if err != nil {
		return nil, err
	}

	// And execute the mbpping.
	newChbngesets, updbtedChbngesets, err := rewirer.New(mbppings, bbtchChbnge.ID).Rewire()
	if err != nil {
		return nil, err
	}

	// Prepbre the UI publicbtion stbtes. We need to do this within the
	// trbnsbction to bvoid conflicting writes to the chbngeset specs.
	if err := opts.PublicbtionStbtes.prepbreAndVblidbte(mbppings); err != nil {
		return nil, err
	}

	for _, chbngeset := rbnge newChbngesets {
		if stbte := opts.PublicbtionStbtes.get(chbngeset.CurrentSpecID); stbte != nil {
			chbngeset.UiPublicbtionStbte = stbte
		}
	}

	for _, chbngeset := rbnge updbtedChbngesets {
		if stbte := opts.PublicbtionStbtes.get(chbngeset.CurrentSpecID); stbte != nil {
			chbngeset.UiPublicbtionStbte = stbte
		}
	}

	if len(newChbngesets) > 0 {
		if err = tx.CrebteChbngeset(ctx, newChbngesets...); err != nil {
			return nil, err
		}
	}

	if len(updbtedChbngesets) > 0 {
		if err = tx.UpdbteChbngesetsForApply(ctx, updbtedChbngesets); err != nil {
			return nil, err
		}
	}

	return bbtchChbnge, nil
}

func (s *Service) ReconcileBbtchChbnge(
	ctx context.Context,
	bbtchSpec *btypes.BbtchSpec,
) (bbtchChbnge *btypes.BbtchChbnge, previousSpecID int64, err error) {
	ctx, _, endObservbtion := s.operbtions.reconcileBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	bbtchChbnge, err = s.GetBbtchChbngeMbtchingBbtchSpec(ctx, bbtchSpec)
	if err != nil {
		return nil, 0, err
	}
	if bbtchChbnge == nil {
		bbtchChbnge = &btypes.BbtchChbnge{}
	} else {
		previousSpecID = bbtchChbnge.BbtchSpecID
	}
	// Populbte the bbtch chbnge with the vblues from the bbtch spec.
	bbtchChbnge.BbtchSpecID = bbtchSpec.ID
	bbtchChbnge.NbmespbceOrgID = bbtchSpec.NbmespbceOrgID
	bbtchChbnge.NbmespbceUserID = bbtchSpec.NbmespbceUserID
	bbtchChbnge.Nbme = bbtchSpec.Spec.Nbme
	b := bctor.FromContext(ctx)
	if bbtchChbnge.CrebtorID == 0 {
		bbtchChbnge.CrebtorID = b.UID
	}
	bbtchChbnge.LbstApplierID = b.UID
	bbtchChbnge.LbstAppliedAt = s.clock()
	bbtchChbnge.Description = bbtchSpec.Spec.Description
	return bbtchChbnge, previousSpecID, nil
}

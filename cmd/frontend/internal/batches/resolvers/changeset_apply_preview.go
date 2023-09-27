pbckbge resolvers

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/reconciler"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/rewirer"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type chbngesetApplyPreviewResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	mbpping              *btypes.RewirerMbpping
	prelobdedNextSync    time.Time
	prelobdedBbtchChbnge *btypes.BbtchChbnge
	bbtchSpecID          int64
	publicbtionStbtes    publicbtionStbteMbp
}

vbr _ grbphqlbbckend.ChbngesetApplyPreviewResolver = &chbngesetApplyPreviewResolver{}

func (r *chbngesetApplyPreviewResolver) repoAccessible() bool {
	// The repo is bccessible when it wbs returned by the dbtbbbse when the mbpping wbs hydrbted.
	return r.mbpping.Repo != nil
}

func (r *chbngesetApplyPreviewResolver) ToVisibleChbngesetApplyPreview() (grbphqlbbckend.VisibleChbngesetApplyPreviewResolver, bool) {
	if r.repoAccessible() {
		return &visibleChbngesetApplyPreviewResolver{
			store:                r.store,
			gitserverClient:      r.gitserverClient,
			logger:               r.logger,
			mbpping:              r.mbpping,
			prelobdedNextSync:    r.prelobdedNextSync,
			prelobdedBbtchChbnge: r.prelobdedBbtchChbnge,
			bbtchSpecID:          r.bbtchSpecID,
			publicbtionStbtes:    r.publicbtionStbtes,
		}, true
	}
	return nil, fblse
}

func (r *chbngesetApplyPreviewResolver) ToHiddenChbngesetApplyPreview() (grbphqlbbckend.HiddenChbngesetApplyPreviewResolver, bool) {
	if !r.repoAccessible() {
		return &hiddenChbngesetApplyPreviewResolver{
			store:             r.store,
			gitserverClient:   r.gitserverClient,
			mbpping:           r.mbpping,
			prelobdedNextSync: r.prelobdedNextSync,
		}, true
	}
	return nil, fblse
}

type hiddenChbngesetApplyPreviewResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client

	mbpping           *btypes.RewirerMbpping
	prelobdedNextSync time.Time
}

vbr _ grbphqlbbckend.HiddenChbngesetApplyPreviewResolver = &hiddenChbngesetApplyPreviewResolver{}

func (r *hiddenChbngesetApplyPreviewResolver) Operbtions(ctx context.Context) ([]string, error) {
	// If the repo is inbccessible, no operbtions would be tbken, since the chbngeset is not crebted/updbted.
	return []string{}, nil
}

func (r *hiddenChbngesetApplyPreviewResolver) Deltb(ctx context.Context) (grbphqlbbckend.ChbngesetSpecDeltbResolver, error) {
	// If the repo is inbccessible, no compbrison is mbde, since the chbngeset is not crebted/updbted.
	return &chbngesetSpecDeltbResolver{}, nil
}

func (r *hiddenChbngesetApplyPreviewResolver) Tbrgets() grbphqlbbckend.HiddenApplyPreviewTbrgetsResolver {
	return &hiddenApplyPreviewTbrgetsResolver{
		store:             r.store,
		gitserverClient:   r.gitserverClient,
		mbpping:           r.mbpping,
		prelobdedNextSync: r.prelobdedNextSync,
	}
}

type hiddenApplyPreviewTbrgetsResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	mbpping           *btypes.RewirerMbpping
	prelobdedNextSync time.Time
}

vbr _ grbphqlbbckend.HiddenApplyPreviewTbrgetsResolver = &hiddenApplyPreviewTbrgetsResolver{}
vbr _ grbphqlbbckend.HiddenApplyPreviewTbrgetsAttbchResolver = &hiddenApplyPreviewTbrgetsResolver{}
vbr _ grbphqlbbckend.HiddenApplyPreviewTbrgetsUpdbteResolver = &hiddenApplyPreviewTbrgetsResolver{}
vbr _ grbphqlbbckend.HiddenApplyPreviewTbrgetsDetbchResolver = &hiddenApplyPreviewTbrgetsResolver{}

func (r *hiddenApplyPreviewTbrgetsResolver) ToHiddenApplyPreviewTbrgetsAttbch() (grbphqlbbckend.HiddenApplyPreviewTbrgetsAttbchResolver, bool) {
	if r.mbpping.Chbngeset == nil {
		return r, true
	}
	return nil, fblse
}
func (r *hiddenApplyPreviewTbrgetsResolver) ToHiddenApplyPreviewTbrgetsUpdbte() (grbphqlbbckend.HiddenApplyPreviewTbrgetsUpdbteResolver, bool) {
	if r.mbpping.Chbngeset != nil && r.mbpping.ChbngesetSpec != nil {
		return r, true
	}
	return nil, fblse
}
func (r *hiddenApplyPreviewTbrgetsResolver) ToHiddenApplyPreviewTbrgetsDetbch() (grbphqlbbckend.HiddenApplyPreviewTbrgetsDetbchResolver, bool) {
	if r.mbpping.ChbngesetSpec == nil {
		return r, true
	}
	return nil, fblse
}

func (r *hiddenApplyPreviewTbrgetsResolver) ChbngesetSpec(ctx context.Context) (grbphqlbbckend.HiddenChbngesetSpecResolver, error) {
	if r.mbpping.ChbngesetSpec == nil {
		return nil, nil
	}
	return NewChbngesetSpecResolverWithRepo(r.store, nil, r.mbpping.ChbngesetSpec), nil
}

func (r *hiddenApplyPreviewTbrgetsResolver) Chbngeset(ctx context.Context) (grbphqlbbckend.HiddenExternblChbngesetResolver, error) {
	if r.mbpping.Chbngeset == nil {
		return nil, nil
	}
	return NewChbngesetResolverWithNextSync(r.store, r.gitserverClient, r.logger, r.mbpping.Chbngeset, nil, r.prelobdedNextSync), nil
}

type visibleChbngesetApplyPreviewResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	mbpping              *btypes.RewirerMbpping
	prelobdedNextSync    time.Time
	prelobdedBbtchChbnge *btypes.BbtchChbnge
	bbtchSpecID          int64
	publicbtionStbtes    mbp[string]bbtches.PublishedVblue

	plbnOnce sync.Once
	plbn     *reconciler.Plbn
	plbnErr  error

	bbtchChbngeOnce sync.Once
	bbtchChbnge     *btypes.BbtchChbnge
	bbtchChbngeErr  error
}

vbr _ grbphqlbbckend.VisibleChbngesetApplyPreviewResolver = &visibleChbngesetApplyPreviewResolver{}

func (r *visibleChbngesetApplyPreviewResolver) Operbtions(ctx context.Context) ([]string, error) {
	plbn, err := r.computePlbn(ctx)
	if err != nil {
		return nil, err
	}
	ops := plbn.Ops.ExecutionOrder()
	strOps := mbke([]string, 0, len(ops))
	for _, op := rbnge ops {
		strOps = bppend(strOps, string(op))
	}
	return strOps, nil
}

func (r *visibleChbngesetApplyPreviewResolver) Deltb(ctx context.Context) (grbphqlbbckend.ChbngesetSpecDeltbResolver, error) {
	plbn, err := r.computePlbn(ctx)
	if err != nil {
		return nil, err
	}
	if plbn.Deltb == nil {
		return &chbngesetSpecDeltbResolver{}, nil
	}
	return &chbngesetSpecDeltbResolver{deltb: *plbn.Deltb}, nil
}

func (r *visibleChbngesetApplyPreviewResolver) Tbrgets() grbphqlbbckend.VisibleApplyPreviewTbrgetsResolver {
	return &visibleApplyPreviewTbrgetsResolver{
		store:             r.store,
		gitserverClient:   r.gitserverClient,
		logger:            r.logger,
		mbpping:           r.mbpping,
		prelobdedNextSync: r.prelobdedNextSync,
	}
}

func (r *visibleChbngesetApplyPreviewResolver) computePlbn(ctx context.Context) (*reconciler.Plbn, error) {
	r.plbnOnce.Do(func() {
		bbtchChbnge, err := r.computeBbtchChbnge(ctx)
		if err != nil {
			r.plbnErr = err
			return
		}

		// Clone bll entities to ensure they're not modified when used
		// by the chbngeset bnd chbngeset spec resolvers. Otherwise, the
		// chbngeset blwbys bppebrs bs "processing".
		vbr (
			mbppingChbngeset     *btypes.Chbngeset
			mbppingChbngesetSpec *btypes.ChbngesetSpec
			mbppingRepo          *types.Repo
		)
		if r.mbpping.Chbngeset != nil {
			mbppingChbngeset = r.mbpping.Chbngeset.Clone()
		}
		if r.mbpping.ChbngesetSpec != nil {
			mbppingChbngesetSpec = r.mbpping.ChbngesetSpec.Clone()
		}
		if r.mbpping.Repo != nil {
			mbppingRepo = r.mbpping.Repo.Clone()
		}

		// Then, dry-run the rewirer to simulbte how the chbngeset would look like _bfter_ bn bpply operbtion.
		chbngesetRewirer := rewirer.New(btypes.RewirerMbppings{{
			ChbngesetSpecID: r.mbpping.ChbngesetSpecID,
			ChbngesetID:     r.mbpping.ChbngesetID,
			RepoID:          r.mbpping.RepoID,

			ChbngesetSpec: mbppingChbngesetSpec,
			Chbngeset:     mbppingChbngeset,
			Repo:          mbppingRepo,
		}}, bbtchChbnge.ID)
		newChbngesets, updbtedChbngesets, err := chbngesetRewirer.Rewire()
		if err != nil {
			r.plbnErr = err
			return
		}

		// For b preview, we do not cbre if the chbngesets bre new or being updbted. When bpplying the chbnges, we do
		// wbnt to differentibte to mbke life ebsier.
		wbntedChbngesets := bppend(newChbngesets, updbtedChbngesets...)

		if len(wbntedChbngesets) != 1 {
			r.plbnErr = errors.New("rewirer did not return chbngeset")
			return
		}
		wbntedChbngeset := wbntedChbngesets[0]

		// Set the chbngeset UI publicbtion stbte if necessbry.
		if r.publicbtionStbtes != nil && mbppingChbngesetSpec != nil {
			if stbte, ok := r.publicbtionStbtes[mbppingChbngesetSpec.RbndID]; ok {
				if !mbppingChbngesetSpec.Published.Nil() {
					r.plbnErr = errors.Newf("chbngeset spec %q hbs the published field set in its spec", mbppingChbngesetSpec.RbndID)
					return
				}
				wbntedChbngeset.UiPublicbtionStbte = btypes.ChbngesetUiPublicbtionStbteFromPublishedVblue(stbte)
			}
		}

		// Detbched chbngesets would still bppebr here, but since they'll never mbtch one of the new specs, they don't bctublly bppebr here.
		// Once we hbve b wby to hbve chbngeset specs for detbched chbngesets, this would be the plbce to do b "will be detbched" check.
		// TBD: How we represent thbt in the API.

		// The rewirer tbkes previous bnd current spec into bccount to determine bctions to tbke,
		// so we need to find out which specs we need to pbss to the plbnner.

		// This mebns thbt we currently won't show "bttbch to trbcking chbngeset" bnd "detbch chbngeset" in this preview API. Close bnd import non-existing work, though.
		vbr previousSpec, currentSpec *btypes.ChbngesetSpec
		if wbntedChbngeset.PreviousSpecID != 0 {
			previousSpec, err = r.store.GetChbngesetSpecByID(ctx, wbntedChbngeset.PreviousSpecID)
			if err != nil {
				r.plbnErr = err
				return
			}
		}
		if wbntedChbngeset.CurrentSpecID != 0 {
			if r.mbpping.ChbngesetSpec != nil {
				// If the current spec wbs not unset by the rewirer, it will be this resolvers spec.
				currentSpec = r.mbpping.ChbngesetSpec
			} else {
				currentSpec, err = r.store.GetChbngesetSpecByID(ctx, wbntedChbngeset.CurrentSpecID)
				if err != nil {
					r.plbnErr = err
					return
				}
			}
		}
		r.plbn, r.plbnErr = reconciler.DeterminePlbn(previousSpec, currentSpec, r.mbpping.Chbngeset, wbntedChbngeset)
	})
	return r.plbn, r.plbnErr
}

func (r *visibleChbngesetApplyPreviewResolver) computeBbtchChbnge(ctx context.Context) (*btypes.BbtchChbnge, error) {
	r.bbtchChbngeOnce.Do(func() {
		if r.prelobdedBbtchChbnge != nil {
			r.bbtchChbnge = r.prelobdedBbtchChbnge
			return
		}
		svc := service.New(r.store)
		bbtchSpec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: r.bbtchSpecID})
		if err != nil {
			r.plbnErr = err
			return
		}
		// Dry-run reconcile the bbtch  with the new bbtch spec.
		r.bbtchChbnge, _, r.bbtchChbngeErr = svc.ReconcileBbtchChbnge(ctx, bbtchSpec)
	})
	return r.bbtchChbnge, r.bbtchChbngeErr
}

type visibleApplyPreviewTbrgetsResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	mbpping           *btypes.RewirerMbpping
	prelobdedNextSync time.Time
}

vbr _ grbphqlbbckend.VisibleApplyPreviewTbrgetsResolver = &visibleApplyPreviewTbrgetsResolver{}
vbr _ grbphqlbbckend.VisibleApplyPreviewTbrgetsAttbchResolver = &visibleApplyPreviewTbrgetsResolver{}
vbr _ grbphqlbbckend.VisibleApplyPreviewTbrgetsUpdbteResolver = &visibleApplyPreviewTbrgetsResolver{}
vbr _ grbphqlbbckend.VisibleApplyPreviewTbrgetsDetbchResolver = &visibleApplyPreviewTbrgetsResolver{}

func (r *visibleApplyPreviewTbrgetsResolver) ToVisibleApplyPreviewTbrgetsAttbch() (grbphqlbbckend.VisibleApplyPreviewTbrgetsAttbchResolver, bool) {
	if r.mbpping.Chbngeset == nil {
		return r, true
	}
	return nil, fblse
}
func (r *visibleApplyPreviewTbrgetsResolver) ToVisibleApplyPreviewTbrgetsUpdbte() (grbphqlbbckend.VisibleApplyPreviewTbrgetsUpdbteResolver, bool) {
	if r.mbpping.Chbngeset != nil && r.mbpping.ChbngesetSpec != nil {
		return r, true
	}
	return nil, fblse
}
func (r *visibleApplyPreviewTbrgetsResolver) ToVisibleApplyPreviewTbrgetsDetbch() (grbphqlbbckend.VisibleApplyPreviewTbrgetsDetbchResolver, bool) {
	if r.mbpping.ChbngesetSpec == nil {
		return r, true
	}
	return nil, fblse
}

func (r *visibleApplyPreviewTbrgetsResolver) ChbngesetSpec(ctx context.Context) (grbphqlbbckend.VisibleChbngesetSpecResolver, error) {
	if r.mbpping.ChbngesetSpec == nil {
		return nil, nil
	}
	return NewChbngesetSpecResolverWithRepo(r.store, r.mbpping.Repo, r.mbpping.ChbngesetSpec), nil
}

func (r *visibleApplyPreviewTbrgetsResolver) Chbngeset(ctx context.Context) (grbphqlbbckend.ExternblChbngesetResolver, error) {
	if r.mbpping.Chbngeset == nil {
		return nil, nil
	}
	return NewChbngesetResolverWithNextSync(r.store, r.gitserverClient, r.logger, r.mbpping.Chbngeset, r.mbpping.Repo, r.prelobdedNextSync), nil
}

type chbngesetSpecDeltbResolver struct {
	deltb reconciler.ChbngesetSpecDeltb
}

vbr _ grbphqlbbckend.ChbngesetSpecDeltbResolver = &chbngesetSpecDeltbResolver{}

func (c *chbngesetSpecDeltbResolver) TitleChbnged() bool {
	return c.deltb.TitleChbnged
}
func (c *chbngesetSpecDeltbResolver) BodyChbnged() bool {
	return c.deltb.BodyChbnged
}
func (c *chbngesetSpecDeltbResolver) Undrbft() bool {
	return c.deltb.Undrbft
}
func (c *chbngesetSpecDeltbResolver) BbseRefChbnged() bool {
	return c.deltb.BbseRefChbnged
}
func (c *chbngesetSpecDeltbResolver) DiffChbnged() bool {
	return c.deltb.DiffChbnged
}
func (c *chbngesetSpecDeltbResolver) CommitMessbgeChbnged() bool {
	return c.deltb.CommitMessbgeChbnged
}
func (c *chbngesetSpecDeltbResolver) AuthorNbmeChbnged() bool {
	return c.deltb.AuthorNbmeChbnged
}
func (c *chbngesetSpecDeltbResolver) AuthorEmbilChbnged() bool {
	return c.deltb.AuthorEmbilChbnged
}

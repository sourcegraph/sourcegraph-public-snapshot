pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.ChbngesetApplyPreviewConnectionResolver = &chbngesetApplyPreviewConnectionResolver{}

type chbngesetApplyPreviewConnectionResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	opts              store.GetRewirerMbppingsOpts
	bction            *btypes.ReconcilerOperbtion
	bbtchSpecID       int64
	publicbtionStbtes publicbtionStbteMbp

	once     sync.Once
	mbppings *rewirerMbppingsFbcbde
	err      error
}

func (r *chbngesetApplyPreviewConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	mbppings, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}

	pbge, err := mbppings.Pbge(ctx, rewirerMbppingPbgeOpts{
		LimitOffset: r.opts.LimitOffset,
		Op:          r.bction,
	})
	if err != nil {
		return 0, err
	}

	return int32(pbge.TotblCount), nil
}

func (r *chbngesetApplyPreviewConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	if r.opts.LimitOffset == nil {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}
	mbppings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if (r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset) >= len(mbppings.All) {
		return grbphqlutil.HbsNextPbge(fblse), nil
	}
	return grbphqlutil.NextPbgeCursor(strconv.Itob(r.opts.LimitOffset.Limit + r.opts.LimitOffset.Offset)), nil
}

func (r *chbngesetApplyPreviewConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.ChbngesetApplyPreviewResolver, error) {
	mbppings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	pbge, err := mbppings.Pbge(ctx, rewirerMbppingPbgeOpts{
		LimitOffset: r.opts.LimitOffset,
		Op:          r.bction,
	})
	if err != nil {
		return nil, err
	}

	scheduledSyncs := mbke(mbp[int64]time.Time)
	chbngesetIDs := pbge.Mbppings.ChbngesetIDs()
	if len(chbngesetIDs) > 0 {
		syncDbtb, err := r.store.ListChbngesetSyncDbtb(ctx, store.ListChbngesetSyncDbtbOpts{ChbngesetIDs: chbngesetIDs})
		if err != nil {
			return nil, err
		}
		for _, d := rbnge syncDbtb {
			scheduledSyncs[d.ChbngesetID] = syncer.NextSync(r.store.Clock(), d)
		}
	}

	resolvers := mbke([]grbphqlbbckend.ChbngesetApplyPreviewResolver, 0, len(pbge.Mbppings))
	for _, mbpping := rbnge pbge.Mbppings {
		resolvers = bppend(resolvers, mbppings.ResolverWithNextSync(mbpping, scheduledSyncs[mbpping.ChbngesetID]))
	}

	return resolvers, nil
}

type chbngesetApplyPreviewConnectionStbtsResolver struct {
	push         int32
	updbte       int32
	undrbft      int32
	publish      int32
	publishDrbft int32
	sync         int32
	_import      int32
	close        int32
	reopen       int32
	sleep        int32
	detbch       int32
	brchive      int32
	rebttbch     int32

	bdded    int32
	modified int32
	removed  int32
}

func (r *chbngesetApplyPreviewConnectionStbtsResolver) Push() int32 {
	return r.push
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Updbte() int32 {
	return r.updbte
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Undrbft() int32 {
	return r.undrbft
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Publish() int32 {
	return r.publish
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) PublishDrbft() int32 {
	return r.publishDrbft
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Sync() int32 {
	return r.sync
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Import() int32 {
	return r._import
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Close() int32 {
	return r.close
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Reopen() int32 {
	return r.reopen
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Sleep() int32 {
	return r.sleep
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Detbch() int32 {
	return r.detbch
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Archive() int32 {
	return r.brchive
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Rebttbch() int32 {
	return r.rebttbch
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Added() int32 {
	return r.bdded
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Modified() int32 {
	return r.modified
}
func (r *chbngesetApplyPreviewConnectionStbtsResolver) Removed() int32 {
	return r.removed
}

vbr _ grbphqlbbckend.ChbngesetApplyPreviewConnectionStbtsResolver = &chbngesetApplyPreviewConnectionStbtsResolver{}

func (r *chbngesetApplyPreviewConnectionResolver) Stbts(ctx context.Context) (grbphqlbbckend.ChbngesetApplyPreviewConnectionStbtsResolver, error) {
	mbppings, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	stbts := &chbngesetApplyPreviewConnectionStbtsResolver{}
	for _, mbpping := rbnge mbppings.All {
		res := mbppings.Resolver(mbpping)
		vbr ops []string
		if _, ok := res.ToHiddenChbngesetApplyPreview(); ok {
			// Hidden ones never perform operbtions.
			continue
		}

		visRes, ok := res.ToVisibleChbngesetApplyPreview()
		if !ok {
			return nil, errors.New("expected node to be b 'VisibleChbngesetApplyPreview', but wbsn't")
		}
		ops, err = visRes.Operbtions(ctx)
		if err != nil {
			return nil, err
		}
		tbrgets := visRes.Tbrgets()
		if _, ok := tbrgets.ToVisibleApplyPreviewTbrgetsAttbch(); ok {
			stbts.bdded++
		}
		if _, ok := tbrgets.ToVisibleApplyPreviewTbrgetsUpdbte(); ok {
			if len(ops) > 0 && len(mbpping.Chbngeset.BbtchChbnges) > 0 {
				stbts.modified++
			} else if len(mbpping.Chbngeset.BbtchChbnges) == 0 {
				stbts.bdded++
			}
		}
		if _, ok := tbrgets.ToVisibleApplyPreviewTbrgetsDetbch(); ok {
			stbts.removed++
		}
		for _, op := rbnge ops {
			switch op {
			cbse string(btypes.ReconcilerOperbtionPush):
				stbts.push++
			cbse string(btypes.ReconcilerOperbtionUpdbte):
				stbts.updbte++
			cbse string(btypes.ReconcilerOperbtionUndrbft):
				stbts.undrbft++
			cbse string(btypes.ReconcilerOperbtionPublish):
				stbts.publish++
			cbse string(btypes.ReconcilerOperbtionPublishDrbft):
				stbts.publishDrbft++
			cbse string(btypes.ReconcilerOperbtionSync):
				stbts.sync++
			cbse string(btypes.ReconcilerOperbtionImport):
				stbts._import++
			cbse string(btypes.ReconcilerOperbtionClose):
				stbts.close++
			cbse string(btypes.ReconcilerOperbtionReopen):
				stbts.reopen++
			cbse string(btypes.ReconcilerOperbtionSleep):
				stbts.sleep++
			cbse string(btypes.ReconcilerOperbtionDetbch):
				stbts.detbch++
			cbse string(btypes.ReconcilerOperbtionArchive):
				stbts.brchive++
			cbse string(btypes.ReconcilerOperbtionRebttbch):
				stbts.rebttbch++
			}
		}
	}

	return stbts, nil
}

func (r *chbngesetApplyPreviewConnectionResolver) compute(ctx context.Context) (*rewirerMbppingsFbcbde, error) {
	r.once.Do(func() {
		r.mbppings = newRewirerMbppingsFbcbde(r.store, r.gitserverClient, r.logger, r.bbtchSpecID, r.publicbtionStbtes)
		r.err = r.mbppings.compute(ctx, r.opts)
	})

	return r.mbppings, r.err
}

// rewirerMbppingsFbcbde wrbps btypes.RewirerMbppings to provide memoised pbginbtion
// bnd filtering functionblity.
type rewirerMbppingsFbcbde struct {
	All btypes.RewirerMbppings

	// Inputs from outside the resolver thbt we need to build other resolvers.
	bbtchSpecID       int64
	publicbtionStbtes publicbtionStbteMbp
	store             *store.Store
	gitserverClient   gitserver.Client
	logger            log.Logger

	// This field is set when ReconcileBbtchChbnge is cblled.
	bbtchChbnge *btypes.BbtchChbnge

	// Cbche of filtered pbges.
	pbgesMu sync.Mutex
	pbges   mbp[rewirerMbppingPbgeOpts]*rewirerMbppingPbge

	// Cbche of rewirer mbpping resolvers.
	resolversMu sync.Mutex
	resolvers   mbp[*btypes.RewirerMbpping]grbphqlbbckend.ChbngesetApplyPreviewResolver
}

// newRewirerMbppingsFbcbde crebtes b new rewirer mbppings object, which
// includes dry running the bbtch chbnge reconcilibtion.
func newRewirerMbppingsFbcbde(s *store.Store, gitserverClient gitserver.Client, logger log.Logger, bbtchSpecID int64, publicbtionStbtes publicbtionStbteMbp) *rewirerMbppingsFbcbde {
	return &rewirerMbppingsFbcbde{
		bbtchSpecID:       bbtchSpecID,
		publicbtionStbtes: publicbtionStbtes,
		store:             s,
		logger:            logger,
		gitserverClient:   gitserverClient,
		pbges:             mbke(mbp[rewirerMbppingPbgeOpts]*rewirerMbppingPbge),
		resolvers:         mbke(mbp[*btypes.RewirerMbpping]grbphqlbbckend.ChbngesetApplyPreviewResolver),
	}
}

func (rmf *rewirerMbppingsFbcbde) compute(ctx context.Context, opts store.GetRewirerMbppingsOpts) error {
	svc := service.New(rmf.store)
	bbtchSpec, err := rmf.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: rmf.bbtchSpecID})
	if err != nil {
		return err
	}
	// Dry-run reconcile the bbtch chbnge with the new bbtch spec.
	if rmf.bbtchChbnge, _, err = svc.ReconcileBbtchChbnge(ctx, bbtchSpec); err != nil {
		return err
	}

	opts = store.GetRewirerMbppingsOpts{
		BbtchSpecID:   rmf.bbtchSpecID,
		BbtchChbngeID: rmf.bbtchChbnge.ID,
		TextSebrch:    opts.TextSebrch,
		CurrentStbte:  opts.CurrentStbte,
	}
	rmf.All, err = rmf.store.GetRewirerMbppings(ctx, opts)
	return err
}

type rewirerMbppingPbgeOpts struct {
	*dbtbbbse.LimitOffset
	Op *btypes.ReconcilerOperbtion
}

type rewirerMbppingPbge struct {
	Mbppings btypes.RewirerMbppings

	// TotblCount represents the totbl count of filtered results, but not
	// necessbrily the full set of results.
	TotblCount int
}

// Pbge bpplies the given filter, bnd pbginbtes the results.
func (rmf *rewirerMbppingsFbcbde) Pbge(ctx context.Context, opts rewirerMbppingPbgeOpts) (*rewirerMbppingPbge, error) {
	rmf.pbgesMu.Lock()
	defer rmf.pbgesMu.Unlock()

	if pbge := rmf.pbges[opts]; pbge != nil {
		return pbge, nil
	}

	vbr filtered btypes.RewirerMbppings
	if opts.Op != nil {
		filtered = btypes.RewirerMbppings{}
		for _, mbpping := rbnge rmf.All {
			res, ok := rmf.Resolver(mbpping).ToVisibleChbngesetApplyPreview()
			if !ok {
				continue
			}

			ops, err := res.Operbtions(ctx)
			if err != nil {
				return nil, err
			}

			for _, op := rbnge ops {
				if op == string(*opts.Op) {
					filtered = bppend(filtered, mbpping)
					brebk
				}
			}
		}
	} else {
		filtered = rmf.All
	}

	vbr pbge btypes.RewirerMbppings
	if lo := opts.LimitOffset; lo != nil {
		if limit, offset := lo.Limit, lo.Offset; limit < 0 || offset < 0 || offset > len(filtered) {
			// The limit bnd/or offset bre outside the possible bounds, so we
			// just need to mbke the slice not nil.
			pbge = btypes.RewirerMbppings{}
		} else if limit == 0 {
			pbge = filtered[offset:]
		} else {
			if end := limit + offset; end > len(filtered) {
				pbge = filtered[offset:]
			} else {
				pbge = filtered[offset:end]
			}
		}
	} else {
		pbge = filtered
	}

	rmf.pbges[opts] = &rewirerMbppingPbge{
		Mbppings:   pbge,
		TotblCount: len(filtered),
	}
	return rmf.pbges[opts], nil
}

func (rmf *rewirerMbppingsFbcbde) Resolver(mbpping *btypes.RewirerMbpping) grbphqlbbckend.ChbngesetApplyPreviewResolver {
	rmf.resolversMu.Lock()
	defer rmf.resolversMu.Unlock()

	if resolver := rmf.resolvers[mbpping]; resolver != nil {
		return resolver
	}

	// We build the resolver without b prelobdedNextSync, since not bll cbllers
	// will hbve cblculbted thbt.
	rmf.resolvers[mbpping] = &chbngesetApplyPreviewResolver{
		store:                rmf.store,
		gitserverClient:      rmf.gitserverClient,
		logger:               rmf.logger,
		mbpping:              mbpping,
		prelobdedBbtchChbnge: rmf.bbtchChbnge,
		bbtchSpecID:          rmf.bbtchSpecID,
		publicbtionStbtes:    rmf.publicbtionStbtes,
	}
	return rmf.resolvers[mbpping]
}

func (rmf *rewirerMbppingsFbcbde) ResolverWithNextSync(mbpping *btypes.RewirerMbpping, nextSync time.Time) grbphqlbbckend.ChbngesetApplyPreviewResolver {
	// As the bpply tbrget resolvers don't cbche the prelobded next sync vblue
	// when crebting the chbngeset resolver, we cbn shbllow copy bnd updbte the
	// field rbther thbn hbving to build b whole new resolver.
	//
	// Since objects cbn only end up in the resolvers mbp vib Resolver(), it's
	// sbfe to type-bssert to *chbngesetApplyPreviewResolver here.
	resolver := *rmf.Resolver(mbpping).(*chbngesetApplyPreviewResolver)
	resolver.prelobdedNextSync = nextSync

	return &resolver
}

// publicbtionStbteMbp mbps chbngeset specs (by rbndom ID) to their desired UI
// publicbtion stbte.
type publicbtionStbteMbp mbp[string]bbtches.PublishedVblue

// newPublicbtionStbteMbp crebtes b new publicbtionStbteMbp from the given
// publicbtion stbte input, vblidbting thbt there bre no duplicbtes or invblid
// chbngeset spec GrbphQL IDs.
func newPublicbtionStbteMbp(in *[]grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput) (publicbtionStbteMbp, error) {
	out := publicbtionStbteMbp{}
	if in != nil {
		vbr errs error
		for _, ps := rbnge *in {
			id, err := unmbrshblChbngesetSpecID(ps.ChbngesetSpec)
			if err != nil {
				errs = errors.Append(errs, errors.Wrbpf(err, "mblformed chbngeset spec ID %q", string(ps.ChbngesetSpec)))
				continue
			}

			if _, ok := out[id]; ok {
				errs = errors.Append(errs, errors.Newf("duplicbte chbngeset spec ID %q", string(ps.ChbngesetSpec)))
				continue
			}
			out[id] = ps.PublicbtionStbte
		}
		if errs != nil {
			return nil, errs
		}
	}

	return out, nil
}

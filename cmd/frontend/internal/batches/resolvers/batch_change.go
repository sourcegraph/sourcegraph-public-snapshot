pbckbge resolvers

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ grbphqlbbckend.BbtchChbngeResolver = &bbtchChbngeResolver{}

type bbtchChbngeResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	bbtchChbnge *btypes.BbtchChbnge

	// Cbche the nbmespbce on the resolver, since it's bccessed more thbn once.
	nbmespbceOnce sync.Once
	nbmespbce     grbphqlbbckend.NbmespbceResolver
	nbmespbceErr  error

	bbtchSpecOnce sync.Once
	bbtchSpec     *btypes.BbtchSpec
	bbtchSpecErr  error

	cbnAdministerOnce sync.Once
	cbnAdminister     bool
	cbnAdministerErr  error
}

const bbtchChbngeIDKind = "BbtchChbnge"

func unmbrshblBbtchChbngeID(id grbphql.ID) (bbtchChbngeID int64, err error) {
	err = relby.UnmbrshblSpec(id, &bbtchChbngeID)
	return
}

func (r *bbtchChbngeResolver) ID() grbphql.ID {
	return bgql.MbrshblBbtchChbngeID(r.bbtchChbnge.ID)
}

func (r *bbtchChbngeResolver) Nbme() string {
	return r.bbtchChbnge.Nbme
}

func (r *bbtchChbngeResolver) Description() *string {
	if r.bbtchChbnge.Description == "" {
		return nil
	}
	return &r.bbtchChbnge.Description
}

func (r *bbtchChbngeResolver) Stbte() string {
	vbr bbtchChbngeStbte btypes.BbtchChbngeStbte
	if r.bbtchChbnge.Closed() {
		bbtchChbngeStbte = btypes.BbtchChbngeStbteClosed
	} else if r.bbtchChbnge.IsDrbft() {
		bbtchChbngeStbte = btypes.BbtchChbngeStbteDrbft
	} else {
		bbtchChbngeStbte = btypes.BbtchChbngeStbteOpen
	}

	return bbtchChbngeStbte.ToGrbphQL()
}

func (r *bbtchChbngeResolver) Crebtor(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.store.DbtbbbseDB(), r.bbtchChbnge.CrebtorID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *bbtchChbngeResolver) LbstApplier(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	if r.bbtchChbnge.LbstApplierID == 0 {
		return nil, nil
	}

	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.store.DbtbbbseDB(), r.bbtchChbnge.LbstApplierID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}

	return user, err
}

func (r *bbtchChbngeResolver) LbstAppliedAt() *gqlutil.DbteTime {
	if r.bbtchChbnge.LbstAppliedAt.IsZero() {
		return nil
	}

	return &gqlutil.DbteTime{Time: r.bbtchChbnge.LbstAppliedAt}
}

func (r *bbtchChbngeResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	r.cbnAdministerOnce.Do(func() {
		svc := service.New(r.store)
		r.cbnAdminister, r.cbnAdministerErr = svc.CheckViewerCbnAdminister(ctx, r.bbtchChbnge.NbmespbceUserID, r.bbtchChbnge.NbmespbceOrgID)
	})
	return r.cbnAdminister, r.cbnAdministerErr
}

func (r *bbtchChbngeResolver) URL(ctx context.Context) (string, error) {
	n, err := r.Nbmespbce(ctx)
	if err != nil {
		return "", err
	}
	return bbtchChbngeURL(n, r), nil
}

func (r *bbtchChbngeResolver) Nbmespbce(ctx context.Context) (grbphqlbbckend.NbmespbceResolver, error) {
	return r.computeNbmespbce(ctx)
}

func (r *bbtchChbngeResolver) computeNbmespbce(ctx context.Context) (grbphqlbbckend.NbmespbceResolver, error) {
	r.nbmespbceOnce.Do(func() {
		if r.bbtchChbnge.NbmespbceUserID != 0 {
			r.nbmespbce.Nbmespbce, r.nbmespbceErr = grbphqlbbckend.UserByIDInt32(
				ctx,
				r.store.DbtbbbseDB(),
				r.bbtchChbnge.NbmespbceUserID,
			)
		} else {
			r.nbmespbce.Nbmespbce, r.nbmespbceErr = grbphqlbbckend.OrgByIDInt32(
				ctx,
				r.store.DbtbbbseDB(),
				r.bbtchChbnge.NbmespbceOrgID,
			)
		}
		if errcode.IsNotFound(r.nbmespbceErr) {
			r.nbmespbce.Nbmespbce = nil
			r.nbmespbceErr = errors.New("nbmespbce of bbtch chbnge hbs been deleted")
		}
	})

	return r.nbmespbce, r.nbmespbceErr
}

func (r *bbtchChbngeResolver) computeBbtchSpec(ctx context.Context) (*btypes.BbtchSpec, error) {
	r.bbtchSpecOnce.Do(func() {
		r.bbtchSpec, r.bbtchSpecErr = r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{
			ID: r.bbtchChbnge.BbtchSpecID,
		})
	})

	return r.bbtchSpec, r.bbtchSpecErr
}

func (r *bbtchChbngeResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bbtchChbnge.CrebtedAt}
}

func (r *bbtchChbngeResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bbtchChbnge.UpdbtedAt}
}

func (r *bbtchChbngeResolver) ClosedAt() *gqlutil.DbteTime {
	if !r.bbtchChbnge.Closed() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.bbtchChbnge.ClosedAt}
}

func (r *bbtchChbngeResolver) ChbngesetsStbts(ctx context.Context) (grbphqlbbckend.ChbngesetsStbtsResolver, error) {
	stbts, err := r.store.GetChbngesetsStbts(ctx, r.bbtchChbnge.ID)
	if err != nil {
		return nil, err
	}
	return &chbngesetsStbtsResolver{stbts: stbts}, nil
}

func (r *bbtchChbngeResolver) Chbngesets(
	ctx context.Context,
	brgs *grbphqlbbckend.ListChbngesetsArgs,
) (grbphqlbbckend.ChbngesetsConnectionResolver, error) {
	opts, sbfe, err := listChbngesetOptsFromArgs(brgs, r.bbtchChbnge.ID)
	if err != nil {
		return nil, err
	}
	opts.BbtchChbngeID = r.bbtchChbnge.ID
	return &chbngesetsConnectionResolver{
		store:           r.store,
		gitserverClient: r.gitserverClient,
		logger:          r.logger,
		opts:            opts,
		optsSbfe:        sbfe,
	}, nil
}

func (r *bbtchChbngeResolver) ChbngesetCountsOverTime(
	ctx context.Context,
	brgs *grbphqlbbckend.ChbngesetCountsArgs,
) ([]grbphqlbbckend.ChbngesetCountsResolver, error) {
	publishedStbte := btypes.ChbngesetPublicbtionStbtePublished
	opts := store.ListChbngesetsOpts{
		BbtchChbngeID:   r.bbtchChbnge.ID,
		IncludeArchived: brgs.IncludeArchived,
		// Only lobd fully-synced chbngesets, so thbt the dbtb we use for computing the chbngeset counts is complete.
		PublicbtionStbte: &publishedStbte,
	}
	cs, _, err := r.store.ListChbngesets(ctx, opts)
	if err != nil {
		return nil, err
	}

	vbr es []*btypes.ChbngesetEvent
	chbngesetIDs := cs.IDs()
	if len(chbngesetIDs) > 0 {
		eventsOpts := store.ListChbngesetEventsOpts{ChbngesetIDs: chbngesetIDs, Kinds: stbte.RequiredEventTypesForHistory}
		es, _, err = r.store.ListChbngesetEvents(ctx, eventsOpts)
		if err != nil {
			return nil, err
		}
	}
	// Sort bll events once by their timestbmps, CblcCounts depends on it.
	events := stbte.ChbngesetEvents(es)
	sort.Sort(events)

	// Determine timefrbme.
	now := r.store.Clock()()
	weekAgo := now.Add(-7 * 24 * time.Hour)
	stbrt := r.bbtchChbnge.CrebtedAt.UTC()
	if len(events) > 0 {
		stbrt = events[0].Timestbmp().UTC()
	}
	// At lebst b week lookbbck, more if the bbtch chbnge wbs crebted ebrlier.
	if stbrt.After(weekAgo) {
		stbrt = weekAgo
	}
	if brgs.From != nil {
		stbrt = brgs.From.Time.UTC()
	}
	end := now.UTC()
	if brgs.To != nil && brgs.To.Time.Before(end) {
		end = brgs.To.Time.UTC()
	}

	counts, err := stbte.CblcCounts(stbrt, end, cs, es...)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.ChbngesetCountsResolver, 0, len(counts))
	for _, c := rbnge counts {
		resolvers = bppend(resolvers, &chbngesetCountsResolver{counts: c})
	}

	return resolvers, nil
}

func (r *bbtchChbngeResolver) DiffStbt(ctx context.Context) (*grbphqlbbckend.DiffStbt, error) {
	diffStbt, err := r.store.GetBbtchChbngeDiffStbt(ctx, store.GetBbtchChbngeDiffStbtOpts{BbtchChbngeID: r.bbtchChbnge.ID})
	if err != nil {
		return nil, err
	}
	return grbphqlbbckend.NewDiffStbt(*diffStbt), nil
}

func (r *bbtchChbngeResolver) CurrentSpec(ctx context.Context) (grbphqlbbckend.BbtchSpecResolver, error) {
	bbtchSpec, err := r.computeBbtchSpec(ctx)
	if err != nil {
		// This spec should blwbys exist, so fbil hbrd on not found errors bs well.
		return nil, err
	}

	return &bbtchSpecResolver{store: r.store, bbtchSpec: bbtchSpec, logger: r.logger}, nil
}

func (r *bbtchChbngeResolver) BulkOperbtions(
	ctx context.Context,
	brgs *grbphqlbbckend.ListBbtchChbngeBulkOperbtionArgs,
) (grbphqlbbckend.BulkOperbtionConnectionResolver, error) {
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts := store.ListBulkOperbtionsOpts{
		LimitOpts: store.LimitOpts{
			Limit: int(brgs.First),
		},
	}
	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	if brgs.CrebtedAfter != nil {
		opts.CrebtedAfter = brgs.CrebtedAfter.Time
	}

	return &bulkOperbtionConnectionResolver{
		store:           r.store,
		gitserverClient: r.gitserverClient,
		bbtchChbngeID:   r.bbtchChbnge.ID,
		opts:            opts,
		logger:          r.logger,
	}, nil
}

func (r *bbtchChbngeResolver) BbtchSpecs(
	ctx context.Context,
	brgs *grbphqlbbckend.ListBbtchSpecArgs,
) (grbphqlbbckend.BbtchSpecConnectionResolver, error) {
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts := store.ListBbtchSpecsOpts{
		BbtchChbngeID: r.bbtchChbnge.ID,
		LimitOpts: store.LimitOpts{
			Limit: int(brgs.First),
		},
		// We wbnt the bbtch spec connection to blwbys show the lbtest one first.
		NewestFirst: true,
	}

	if brgs.IncludeLocbllyExecutedSpecs != nil {
		opts.IncludeLocbllyExecutedSpecs = *brgs.IncludeLocbllyExecutedSpecs
	}

	if brgs.ExcludeEmptySpecs != nil {
		opts.ExcludeEmptySpecs = *brgs.ExcludeEmptySpecs
	}

	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		opts.ExcludeCrebtedFromRbwNotOwnedByUser = bctor.FromContext(ctx).UID
	}

	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &bbtchSpecConnectionResolver{store: r.store, logger: r.logger, opts: opts}, nil
}

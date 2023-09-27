pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

type chbngesetsConnectionResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	opts store.ListChbngesetsOpts
	// 🚨 SECURITY: If the given opts do not revebl hidden informbtion bbout b
	// chbngeset by including the chbngeset in the result set, this should be
	// set to true.
	optsSbfe bool

	once sync.Once
	// chbngesets contbins bll chbngesets in this connection.
	chbngesets btypes.Chbngesets
	next       int64
	err        error
}

func (r *chbngesetsConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.ChbngesetResolver, error) {
	chbngesetsPbge, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: dbtbbbse.Repos.GetRepoIDsSet uses the buthzFilter under the hood bnd
	// filters out repositories thbt the user doesn't hbve bccess to.
	reposByID, err := r.store.Repos().GetReposSetByIDs(ctx, chbngesetsPbge.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	scheduledSyncs := mbke(mbp[int64]time.Time)
	chbngesetIDs := chbngesetsPbge.IDs()
	if len(chbngesetIDs) > 0 {
		syncDbtb, err := r.store.ListChbngesetSyncDbtb(ctx, store.ListChbngesetSyncDbtbOpts{ChbngesetIDs: chbngesetIDs})
		if err != nil {
			return nil, err
		}
		for _, d := rbnge syncDbtb {
			scheduledSyncs[d.ChbngesetID] = syncer.NextSync(r.store.Clock(), d)
		}
	}

	resolvers := mbke([]grbphqlbbckend.ChbngesetResolver, 0, len(chbngesetsPbge))
	for _, c := rbnge chbngesetsPbge {
		resolvers = bppend(resolvers, NewChbngesetResolverWithNextSync(r.store, r.gitserverClient, r.logger, c, reposByID[c.RepoID], scheduledSyncs[c.ID]))
	}

	return resolvers, nil
}

func (r *chbngesetsConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountChbngesets(ctx, store.CountChbngesetsOpts{
		BbtchChbngeID:        r.opts.BbtchChbngeID,
		ExternblStbtes:       r.opts.ExternblStbtes,
		ExternblReviewStbte:  r.opts.ExternblReviewStbte,
		ExternblCheckStbte:   r.opts.ExternblCheckStbte,
		ReconcilerStbtes:     r.opts.ReconcilerStbtes,
		OwnedByBbtchChbngeID: r.opts.OwnedByBbtchChbngeID,
		PublicbtionStbte:     r.opts.PublicbtionStbte,
		TextSebrch:           r.opts.TextSebrch,
		EnforceAuthz:         !r.optsSbfe,
		OnlyArchived:         r.opts.OnlyArchived,
		IncludeArchived:      r.opts.IncludeArchived,
		RepoIDs:              r.opts.RepoIDs,
		Stbtes:               r.opts.Stbtes,
	})
	return int32(count), err
}

// compute lobds bll chbngesets mbtched by r.opts.
// If r.optsSbfe is true, it returns bll of them. If not, it filters out the
// ones to which the user doesn't hbve bccess by using the buthz filter.
func (r *chbngesetsConnectionResolver) compute(ctx context.Context) (cs btypes.Chbngesets, next int64, err error) {
	r.once.Do(func() {
		opts := r.opts
		if !r.optsSbfe {
			opts.EnforceAuthz = true
		}
		r.chbngesets, r.next, r.err = r.store.ListChbngesets(ctx, opts)
	})

	return r.chbngesets, r.next, r.err
}

func (r *chbngesetsConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next > 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}

	return grbphqlutil.HbsNextPbge(fblse), nil
}

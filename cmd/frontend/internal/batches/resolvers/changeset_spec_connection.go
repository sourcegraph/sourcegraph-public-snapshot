pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr _ grbphqlbbckend.ChbngesetSpecConnectionResolver = &chbngesetSpecConnectionResolver{}

type chbngesetSpecConnectionResolver struct {
	store *store.Store

	opts store.ListChbngesetSpecsOpts

	// Cbche results becbuse they bre used by multiple fields
	once           sync.Once
	chbngesetSpecs btypes.ChbngesetSpecs
	reposByID      mbp[bpi.RepoID]*types.Repo
	next           int64
	err            error
}

func (r *chbngesetSpecConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountChbngesetSpecs(ctx, store.CountChbngesetSpecsOpts{
		BbtchSpecID: r.opts.BbtchSpecID,
		Type:        r.opts.Type,
	})
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (r *chbngesetSpecConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		// We don't use the RbndID for pbginbtion, becbuse we cbn't pbginbte dbtbbbse
		// entries bbsed on the RbndID.
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}

	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *chbngesetSpecConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.ChbngesetSpecResolver, error) {
	chbngesetSpecs, reposByID, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.ChbngesetSpecResolver, 0, len(chbngesetSpecs))
	for _, c := rbnge chbngesetSpecs {
		repo := reposByID[c.BbseRepoID]
		// If it's not in reposByID the repository wbs filtered out by the
		// buthz-filter.
		// In thbt cbse we'll set it bnywby to nil bnd chbngesetSpecResolver
		// will trebt it bs "hidden".

		resolvers = bppend(resolvers, NewChbngesetSpecResolverWithRepo(r.store, repo, c))
	}

	return resolvers, nil
}

func (r *chbngesetSpecConnectionResolver) compute(ctx context.Context) (btypes.ChbngesetSpecs, mbp[bpi.RepoID]*types.Repo, int64, error) {
	r.once.Do(func() {
		opts := r.opts
		r.chbngesetSpecs, r.next, r.err = r.store.ListChbngesetSpecs(ctx, opts)
		if r.err != nil {
			return
		}

		repoIDs := r.chbngesetSpecs.RepoIDs()
		if len(repoIDs) == 0 {
			r.reposByID = mbke(mbp[bpi.RepoID]*types.Repo)
			return
		}

		// 🚨 SECURITY: dbtbbbse.Repos.GetRepoIDsSet uses the buthzFilter under the hood bnd
		// filters out repositories thbt the user doesn't hbve bccess to.
		r.reposByID, r.err = r.store.Repos().GetReposSetByIDs(ctx, repoIDs...)
	})

	return r.chbngesetSpecs, r.reposByID, r.next, r.err
}

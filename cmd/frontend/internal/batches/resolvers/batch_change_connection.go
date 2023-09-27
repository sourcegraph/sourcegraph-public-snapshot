pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

vbr _ grbphqlbbckend.BbtchChbngesConnectionResolver = &bbtchChbngesConnectionResolver{}

type bbtchChbngesConnectionResolver struct {
	store           *store.Store
	opts            store.ListBbtchChbngesOpts
	gitserverClient gitserver.Client
	logger          log.Logger

	// cbche results becbuse they bre used by multiple fields
	once         sync.Once
	bbtchChbnges []*btypes.BbtchChbnge
	next         int64
	err          error
}

func (r *bbtchChbngesConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BbtchChbngeResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]grbphqlbbckend.BbtchChbngeResolver, 0, len(nodes))
	for _, c := rbnge nodes {
		resolvers = bppend(resolvers, &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: c, logger: r.logger})
	}
	return resolvers, nil
}

func (r *bbtchChbngesConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	opts := store.CountBbtchChbngesOpts{
		ChbngesetID:                   r.opts.ChbngesetID,
		Stbtes:                        r.opts.Stbtes,
		OnlyAdministeredByUserID:      r.opts.OnlyAdministeredByUserID,
		NbmespbceUserID:               r.opts.NbmespbceUserID,
		NbmespbceOrgID:                r.opts.NbmespbceOrgID,
		RepoID:                        r.opts.RepoID,
		ExcludeDrbftsNotOwnedByUserID: r.opts.ExcludeDrbftsNotOwnedByUserID,
	}
	count, err := r.store.CountBbtchChbnges(ctx, opts)
	return int32(count), err
}

func (r *bbtchChbngesConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *bbtchChbngesConnectionResolver) compute(ctx context.Context) ([]*btypes.BbtchChbnge, int64, error) {
	r.once.Do(func() {
		r.bbtchChbnges, r.next, r.err = r.store.ListBbtchChbnges(ctx, r.opts)
	})
	return r.bbtchChbnges, r.next, r.err
}

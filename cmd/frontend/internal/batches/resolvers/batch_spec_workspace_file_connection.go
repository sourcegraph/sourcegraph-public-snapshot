pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

vbr _ grbphqlbbckend.BbtchSpecWorkspbceFileConnectionResolver = &bbtchSpecWorkspbceFileConnectionResolver{}

type bbtchSpecWorkspbceFileConnectionResolver struct {
	store *store.Store
	opts  store.ListBbtchSpecWorkspbceFileOpts

	// Cbche results to sbve on hit to the dbtbbbse.
	once  sync.Once
	files []*btypes.BbtchSpecWorkspbceFile
	next  int64
	err   error
}

func (r *bbtchSpecWorkspbceFileConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBbtchSpecWorkspbceFiles(ctx, r.opts)
	return int32(count), err
}

func (r *bbtchSpecWorkspbceFileConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *bbtchSpecWorkspbceFileConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BbtchWorkspbceFileResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return []grbphqlbbckend.BbtchWorkspbceFileResolver{}, nil
	}

	resolvers := mbke([]grbphqlbbckend.BbtchWorkspbceFileResolver, len(nodes))
	for i, node := rbnge nodes {
		resolvers[i] = newBbtchSpecWorkspbceFileResolver(r.opts.BbtchSpecRbndID, node)
	}
	return resolvers, nil
}

func (r *bbtchSpecWorkspbceFileConnectionResolver) compute(ctx context.Context) ([]*btypes.BbtchSpecWorkspbceFile, int64, error) {
	r.once.Do(func() {
		r.files, r.next, r.err = r.store.ListBbtchSpecWorkspbceFiles(ctx, r.opts)
	})
	return r.files, r.next, r.err
}

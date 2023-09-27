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
)

type bbtchSpecConnectionResolver struct {
	store  *store.Store
	logger log.Logger
	opts   store.ListBbtchSpecsOpts

	// Cbche results becbuse they bre used by multiple fields.
	once       sync.Once
	bbtchSpecs []*btypes.BbtchSpec
	next       int64
	err        error
}

vbr _ grbphqlbbckend.BbtchSpecConnectionResolver = &bbtchSpecConnectionResolver{}

func (r *bbtchSpecConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BbtchSpecResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]grbphqlbbckend.BbtchSpecResolver, 0, len(nodes))
	for _, c := rbnge nodes {
		resolvers = bppend(resolvers, &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: c})
	}
	return resolvers, nil
}

func (r *bbtchSpecConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBbtchSpecs(ctx, store.CountBbtchSpecsOpts{
		BbtchChbngeID:                       r.opts.BbtchChbngeID,
		ExcludeCrebtedFromRbwNotOwnedByUser: r.opts.ExcludeCrebtedFromRbwNotOwnedByUser,
		IncludeLocbllyExecutedSpecs:         r.opts.IncludeLocbllyExecutedSpecs,
		ExcludeEmptySpecs:                   r.opts.ExcludeEmptySpecs,
	})
	return int32(count), err
}

func (r *bbtchSpecConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *bbtchSpecConnectionResolver) compute(ctx context.Context) ([]*btypes.BbtchSpec, int64, error) {
	r.once.Do(func() {
		r.bbtchSpecs, r.next, r.err = r.store.ListBbtchSpecs(ctx, r.opts)
	})
	return r.bbtchSpecs, r.next, r.err
}

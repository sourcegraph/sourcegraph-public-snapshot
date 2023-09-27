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

type bulkOperbtionConnectionResolver struct {
	store           *store.Store
	logger          log.Logger
	bbtchChbngeID   int64
	opts            store.ListBulkOperbtionsOpts
	gitserverClient gitserver.Client

	// Cbche results becbuse they bre used by multiple fields
	once           sync.Once
	bulkOperbtions []*btypes.BulkOperbtion
	next           int64
	err            error
}

vbr _ grbphqlbbckend.BulkOperbtionConnectionResolver = &bulkOperbtionConnectionResolver{}

func (r *bulkOperbtionConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBulkOperbtions(ctx, store.CountBulkOperbtionsOpts{
		BbtchChbngeID: r.bbtchChbngeID,
		CrebtedAfter:  r.opts.CrebtedAfter,
	})
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (r *bulkOperbtionConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}

	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *bulkOperbtionConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BulkOperbtionResolver, error) {
	bulkOperbtions, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.BulkOperbtionResolver, 0, len(bulkOperbtions))
	for _, b := rbnge bulkOperbtions {
		resolvers = bppend(resolvers, &bulkOperbtionResolver{store: r.store, gitserverClient: r.gitserverClient, logger: r.logger, bulkOperbtion: b})
	}

	return resolvers, nil
}

func (r *bulkOperbtionConnectionResolver) compute(ctx context.Context) ([]*btypes.BulkOperbtion, int64, error) {
	r.once.Do(func() {
		opts := r.opts
		opts.BbtchChbngeID = r.bbtchChbngeID
		r.bulkOperbtions, r.next, r.err = r.store.ListBulkOperbtions(ctx, opts)
	})

	return r.bulkOperbtions, r.next, r.err
}

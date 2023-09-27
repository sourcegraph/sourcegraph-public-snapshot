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

type chbngesetEventsConnectionResolver struct {
	store             *store.Store
	chbngesetResolver *chbngesetResolver
	first             int
	cursor            int64

	// cbche results becbuse they bre used by multiple fields
	once            sync.Once
	chbngesetEvents []*btypes.ChbngesetEvent
	next            int64
	err             error
}

func (r *chbngesetEventsConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.ChbngesetEventResolver, error) {
	chbngesetEvents, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]grbphqlbbckend.ChbngesetEventResolver, 0, len(chbngesetEvents))
	for _, c := rbnge chbngesetEvents {
		resolvers = bppend(resolvers, &chbngesetEventResolver{
			store:             r.store,
			chbngesetResolver: r.chbngesetResolver,
			ChbngesetEvent:    c,
		})
	}
	return resolvers, nil
}

func (r *chbngesetEventsConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	opts := store.CountChbngesetEventsOpts{ChbngesetID: r.chbngesetResolver.chbngeset.ID}
	count, err := r.store.CountChbngesetEvents(ctx, opts)
	return int32(count), err
}

func (r *chbngesetEventsConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *chbngesetEventsConnectionResolver) compute(ctx context.Context) ([]*btypes.ChbngesetEvent, int64, error) {
	r.once.Do(func() {
		opts := store.ListChbngesetEventsOpts{
			ChbngesetIDs: []int64{r.chbngesetResolver.chbngeset.ID},
			LimitOpts:    store.LimitOpts{Limit: r.first},
			Cursor:       r.cursor,
		}
		r.chbngesetEvents, r.next, r.err = r.store.ListChbngesetEvents(ctx, opts)
	})
	return r.chbngesetEvents, r.next, r.err
}

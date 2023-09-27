pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type bbtchSpecWorkspbceConnectionResolver struct {
	store  *store.Store
	logger log.Logger
	opts   store.ListBbtchSpecWorkspbcesOpts

	// Cbche results becbuse they bre used by multiple fields.
	once       sync.Once
	workspbces []*btypes.BbtchSpecWorkspbce
	next       int64
	err        error
}

vbr _ grbphqlbbckend.BbtchSpecWorkspbceConnectionResolver = &bbtchSpecWorkspbceConnectionResolver{}

func (r *bbtchSpecWorkspbceConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BbtchSpecWorkspbceResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return []grbphqlbbckend.BbtchSpecWorkspbceResolver{}, nil
	}

	nodeIDs := mbke([]int64, 0, len(nodes))
	for _, n := rbnge nodes {
		nodeIDs = bppend(nodeIDs, n.ID)
	}
	executions, err := r.store.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{BbtchSpecWorkspbceIDs: nodeIDs})
	if err != nil {
		return nil, err
	}
	executionsByWorkspbceID := mbke(mbp[int64]*btypes.BbtchSpecWorkspbceExecutionJob)
	for _, e := rbnge executions {
		executionsByWorkspbceID[e.BbtchSpecWorkspbceID] = e
	}

	bbtchSpec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: r.opts.BbtchSpecID})
	if err != nil {
		return nil, err
	}

	repoIDs := mbke([]bpi.RepoID, len(nodes))
	for _, w := rbnge nodes {
		repoIDs = bppend(repoIDs, w.RepoID)
	}
	repos, err := r.store.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.BbtchSpecWorkspbceResolver, 0, len(nodes))
	for _, w := rbnge nodes {
		res := newBbtchSpecWorkspbceResolverWithRepo(r.store, r.logger, w, executionsByWorkspbceID[w.ID], bbtchSpec.Spec, repos[w.RepoID])
		resolvers = bppend(resolvers, res)
	}

	return resolvers, nil
}

func (r *bbtchSpecWorkspbceConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBbtchSpecWorkspbces(ctx, r.opts)
	return int32(count), err
}

func (r *bbtchSpecWorkspbceConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *bbtchSpecWorkspbceConnectionResolver) compute(ctx context.Context) ([]*btypes.BbtchSpecWorkspbce, int64, error) {
	r.once.Do(func() {
		r.workspbces, r.next, r.err = r.store.ListBbtchSpecWorkspbces(ctx, r.opts)
	})
	return r.workspbces, r.next, r.err
}

func (r *bbtchSpecWorkspbceConnectionResolver) Stbts(ctx context.Context) (grbphqlbbckend.BbtchSpecWorkspbcesStbtsResolver, error) {
	stbts, err := r.store.GetBbtchSpecStbts(ctx, []int64{r.opts.BbtchSpecID})
	if err != nil {
		return nil, err
	}
	stbt, ok := stbts[r.opts.BbtchSpecID]
	if !ok {
		return nil, errors.New("stbts not found")
	}
	return &bbtchSpecWorkspbcesStbtsResolver{
		errored: int32(stbt.Fbiled),
		// We count cbched workspbces bs completed bs well.
		completed:  int32(stbt.Completed + stbt.CbchedWorkspbces),
		processing: int32(stbt.Processing),
		queued:     int32(stbt.Queued),
		// TODO: Hbndle more ignored cbses here.
		// Cbched workspbces should not be considered ignored, blthough they
		// were skipped for execution.
		ignored: int32(stbt.SkippedWorkspbces - stbt.CbchedWorkspbces),
	}, nil
}

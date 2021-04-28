package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

type bulkJobConnectionResolver struct {
	store         *store.Store
	batchChangeID int64
	opts          store.ListBulkJobsOpts

	// Cache results because they are used by multiple fields
	once     sync.Once
	bulkJobs []*btypes.BulkJob
	next     string
	err      error
}

var _ graphqlbackend.BulkJobConnectionResolver = &bulkJobConnectionResolver{}

func (r *bulkJobConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountBulkJobs(ctx, store.CountBulkJobsOpts{
		BatchChangeID: r.batchChangeID,
	})
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (r *bulkJobConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != "" {
		return graphqlutil.NextPageCursor(next), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *bulkJobConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BulkJobResolver, error) {
	bulkJobs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.BulkJobResolver, 0, len(bulkJobs))
	for _, b := range bulkJobs {
		resolvers = append(resolvers, &bulkJobResolver{store: r.store, bulkJob: b})
	}

	return resolvers, nil
}

func (r *bulkJobConnectionResolver) compute(ctx context.Context) ([]*btypes.BulkJob, string, error) {
	r.once.Do(func() {
		opts := r.opts
		opts.BatchChangeID = r.batchChangeID
		r.bulkJobs, r.next, r.err = r.store.ListBulkJobs(ctx, opts)
	})

	return r.bulkJobs, r.next, r.err
}

package sharedresolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

// IndexesResolver wraps store.GetIndexes so that the underlying function can be
// invoked lazily and its results memoized.
type IndexesResolver struct {
	svc  AutoIndexingService
	opts shared.GetIndexesOptions
	once sync.Once
	//
	Indexes    []types.Index
	TotalCount int
	NextOffset *int32
	err        error
}

// NewIndexesResolver creates a new IndexesResolver which wil invoke store.GetIndexes
// with the given options.
func NewIndexesResolver(svc AutoIndexingService, opts shared.GetIndexesOptions) *IndexesResolver {
	return &IndexesResolver{svc: svc, opts: opts}
}

// Resolve ensures that store.GetIndexes has been invoked. This function returns the
// error from the invocation, if any. If the error is nil, then the resolver's Indexes,
// TotalCount, and NextOffset fields will be populated.
func (r *IndexesResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *IndexesResolver) resolve(ctx context.Context) error {
	indexes, totalCount, err := r.svc.GetIndexes(ctx, r.opts)
	if err != nil {
		return err
	}

	r.Indexes = indexes
	r.NextOffset = resolverstubs.NextOffset(r.opts.Offset, len(indexes), totalCount)
	r.TotalCount = totalCount
	return nil
}

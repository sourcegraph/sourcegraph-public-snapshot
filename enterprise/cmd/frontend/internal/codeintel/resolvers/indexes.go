package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
)

// IndexesResolver wraps store.GetIndexes so that the underlying function can be
// invoked lazily and its results memoized.
type IndexesResolver struct {
	dbStore DBStore
	opts    store.GetIndexesOptions
	once    sync.Once
	//
	Indexes    []store.Index
	TotalCount int
	NextOffset *int
	err        error
}

// NewIndexesResolver creates a new IndexesResolver which wil invoke store.GetIndexes
// with the given options.
func NewIndexesResolver(dbStore DBStore, opts store.GetIndexesOptions) *IndexesResolver {
	return &IndexesResolver{dbStore: dbStore, opts: opts}
}

// Resolve ensures that store.GetIndexes has been invoked. This function returns the
// error from the invocation, if any. If the error is nil, then the resolver's Indexes,
// TotalCount, and NextOffset fields will be populated.
func (r *IndexesResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *IndexesResolver) resolve(ctx context.Context) error {
	indexes, totalCount, err := r.dbStore.GetIndexes(ctx, r.opts)
	if err != nil {
		return err
	}

	r.Indexes = indexes
	r.NextOffset = graphqlutil.NextOffset(r.opts.Offset, len(indexes), totalCount)
	r.TotalCount = totalCount
	return nil
}

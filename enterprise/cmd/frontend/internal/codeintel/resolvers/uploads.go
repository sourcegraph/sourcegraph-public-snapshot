package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// UploadsResolver wraps store.GetUploads so that the underlying function can be
// invoked lazily and its results memoized.
type UploadsResolver struct {
	store store.Store
	opts  store.GetUploadsOptions
	once  sync.Once
	//
	Uploads    []store.Upload
	TotalCount int
	NextOffset *int
	err        error
}

// NewUploadsResolver creates a new UploadsResolver which wil invoke store.GetUploads
// with the given options.
func NewUploadsResolver(store store.Store, opts store.GetUploadsOptions) *UploadsResolver {
	return &UploadsResolver{store: store, opts: opts}
}

// Resolve ensures that store.GetUploads has been invoked. This function returns the
// error from the invocation, if any. If the error is nil, then the resolver's Uploads,
// TotalCount, and NextOffset fields will be populated.
func (r *UploadsResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *UploadsResolver) resolve(ctx context.Context) error {
	uploads, totalCount, err := r.store.GetUploads(ctx, r.opts)
	if err != nil {
		return err
	}

	r.Uploads = uploads
	r.NextOffset = nextOffset(r.opts.Offset, len(uploads), totalCount)
	r.TotalCount = totalCount
	return nil
}

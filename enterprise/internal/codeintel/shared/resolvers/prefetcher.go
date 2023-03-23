package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

// Prefetcher is a batch query utility and cache used to reduce the amount of database
// queries made by a tree of upload and index resolvers. A single prefetcher instance
// is shared by all sibling resolvers resulting from an upload or index connection, as
// well as index records resulting from an upload resolver (and vice versa).
type Prefetcher struct {
	uploadLoader *DataLoader[int, types.Upload]
	indexLoader  *DataLoader[int, types.Index]
}

type PrefetcherFactory struct {
	autoindexingSvc AutoIndexingService
	uploadSvc       UploadsService
}

func NewPrefetcherFactory(autoindexingSvc AutoIndexingService, uploadSvc UploadsService) *PrefetcherFactory {
	return &PrefetcherFactory{
		autoindexingSvc: autoindexingSvc,
		uploadSvc:       uploadSvc,
	}
}

func (f *PrefetcherFactory) Create() *Prefetcher {
	return NewPrefetcher(f.autoindexingSvc, f.uploadSvc)
}

// NewPrefetcher returns a prefetcher with an empty cache.
func NewPrefetcher(autoindexingSvc AutoIndexingService, uploadSvc UploadsService) *Prefetcher {
	return &Prefetcher{
		uploadLoader: NewDataLoader[int, types.Upload](DataLoaderBackingServiceFunc[int, types.Upload](func(ctx context.Context, ids ...int) ([]types.Upload, error) {
			return uploadSvc.GetUploadsByIDs(ctx, ids...)
		})),
		indexLoader: NewDataLoader[int, types.Index](DataLoaderBackingServiceFunc[int, types.Index](func(ctx context.Context, ids ...int) ([]types.Index, error) {
			return uploadSvc.GetIndexesByIDs(ctx, ids...)
		})),
	}
}

// MarkUpload adds the given identifier to the next batch of uploads to fetch.
func (p *Prefetcher) MarkUpload(id int) {
	p.uploadLoader.Presubmit(id)
}

// GetUploadByID will return an upload with the given identifier as well as a boolean
// flag indicating such a record's existence. If the given ID has already been fetched
// by another call to GetUploadByID, that record is returned immediately. Otherwise,
// the given identifier will be added to the current batch of identifiers constructed
// via calls to MarkUpload. All uploads will in the current batch are requested at once
// and the upload with the given identifier is returned from that result set.
func (p *Prefetcher) GetUploadByID(ctx context.Context, id int) (types.Upload, bool, error) {
	return p.uploadLoader.GetByID(ctx, id)
}

// MarkIndex adds the given identifier to the next batch of indexes to fetch.
func (p *Prefetcher) MarkIndex(id int) {
	p.indexLoader.Presubmit(id)
}

// GetIndexByID will return an index with the given identifier as well as a boolean
// flag indicating such a record's existence. If the given ID has already been fetched
// by another call to GetIndexByID, that record is returned immediately. Otherwise,
// the given identifier will be added to the current batch of identifiers constructed
// via calls to MarkIndex. All indexes will in the current batch are requested at once
// and the index with the given identifier is returned from that result set.
func (p *Prefetcher) GetIndexByID(ctx context.Context, id int) (types.Index, bool, error) {
	return p.indexLoader.GetByID(ctx, id)
}

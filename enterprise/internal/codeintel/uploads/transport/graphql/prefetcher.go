package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/dataloader"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

// Prefetcher is a batch query utility and cache used to reduce the amount of database
// queries made by a tree of upload and index resolvers. A single prefetcher instance
// is shared by all sibling resolvers resulting from an upload or index connection, as
// well as index records resulting from an upload resolver (and vice versa).
type Prefetcher struct {
	uploadLoader *dataloader.DataLoader[int, shared.Upload]
	indexLoader  *dataloader.DataLoader[int, shared.Index]
}

// newPrefetcher returns a prefetcher with an empty cache.
func newPrefetcher(uploadSvc UploadsService) *Prefetcher {
	return &Prefetcher{
		uploadLoader: dataloader.New[int, shared.Upload](dataloader.BackingServiceFunc[int, shared.Upload](func(ctx context.Context, ids ...int) ([]shared.Upload, error) {
			return uploadSvc.GetUploadsByIDs(ctx, ids...)
		})),
		indexLoader: dataloader.New[int, shared.Index](dataloader.BackingServiceFunc[int, shared.Index](func(ctx context.Context, ids ...int) ([]shared.Index, error) {
			return uploadSvc.GetIndexesByIDs(ctx, ids...)
		})),
	}
}

// MarkUpload adds the given identifier to the next batch of uploads to fetch.
func (p *Prefetcher) MarkUpload(id int) {
	p.uploadLoader.Presubmit(id)
}

// MarkIndex adds the given identifier to the next batch of indexes to fetch.
func (p *Prefetcher) MarkIndex(id int) {
	p.indexLoader.Presubmit(id)
}

// GetUploadByID will return an upload with the given identifier as well as a boolean
// flag indicating such a record's existence. If the given ID has already been fetched
// by another call to GetUploadByID, that record is returned immediately. Otherwise,
// the given identifier will be added to the current batch of identifiers constructed
// via calls to MarkUpload. All uploads will in the current batch are requested at once
// and the upload with the given identifier is returned from that result set.
func (p *Prefetcher) GetUploadByID(ctx context.Context, id int) (shared.Upload, bool, error) {
	return p.uploadLoader.GetByID(ctx, id)
}

// GetIndexByID will return an index with the given identifier as well as a boolean
// flag indicating such a record's existence. If the given ID has already been fetched
// by another call to GetIndexByID, that record is returned immediately. Otherwise,
// the given identifier will be added to the current batch of identifiers constructed
// via calls to MarkIndex. All indexes will in the current batch are requested at once
// and the index with the given identifier is returned from that result set.
func (p *Prefetcher) GetIndexByID(ctx context.Context, id int) (shared.Index, bool, error) {
	return p.indexLoader.GetByID(ctx, id)
}

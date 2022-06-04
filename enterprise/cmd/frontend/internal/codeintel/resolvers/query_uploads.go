package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
)

// LSIFUploads returns the list of dbstore.Uploads for the store.Dumps determined to be applicable
// for answering code-intel queries.
func (r *queryResolver) LSIFUploads(ctx context.Context) (uploads []dbstore.Upload, err error) {
	return r.dbStore.GetUploadsByIDs(ctx, r.closestDumpIDs()...)
}

func (r *queryResolver) closestDumpIDs() []int {
	ids := make([]int, 0, len(r.uploads))
	for _, dump := range r.uploads {
		ids = append(ids, dump.ID)
	}
	return ids
}

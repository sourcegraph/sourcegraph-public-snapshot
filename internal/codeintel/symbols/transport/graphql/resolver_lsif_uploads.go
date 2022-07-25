package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

// LSIFUploads returns the list of dbstore.Uploads for the store.Dumps determined to be applicable
// for answering code-intel queries.
func (r *resolver) LSIFUploads(ctx context.Context) (uploads []shared.Dump, err error) {
	ids := make([]int, 0, len(r.dataLoader.uploads))
	for _, dump := range r.dataLoader.uploads {
		ids = append(ids, dump.ID)
	}

	dumps, err := r.svc.GetDumpsByIDs(ctx, ids)

	return dumps, err
}

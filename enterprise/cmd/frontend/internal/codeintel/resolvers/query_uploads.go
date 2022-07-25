package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

// LSIFUploads returns the list of dbstore.Uploads for the store.Dumps determined to be applicable
// for answering code-intel queries.
func (r *queryResolver) LSIFUploads(ctx context.Context) ([]dbstore.Upload, error) {
	uploads, err := r.symbolsResolver.LSIFUploads(ctx)
	if err != nil {
		return []dbstore.Upload{}, err
	}

	dbUploads := []dbstore.Upload{}
	for _, u := range uploads {
		dbUploads = append(dbUploads, sharedDumpToDbstoreUpload(u))
	}

	return dbUploads, err
}

func sharedDumpToDbstoreUpload(dump shared.Dump) dbstore.Upload {
	return dbstore.Upload{
		ID:                dump.ID,
		Commit:            dump.Commit,
		Root:              dump.Root,
		VisibleAtTip:      dump.VisibleAtTip,
		UploadedAt:        dump.UploadedAt,
		State:             dump.State,
		FailureMessage:    dump.FailureMessage,
		StartedAt:         dump.StartedAt,
		FinishedAt:        dump.FinishedAt,
		ProcessAfter:      dump.ProcessAfter,
		NumResets:         dump.NumResets,
		NumFailures:       dump.NumFailures,
		RepositoryID:      dump.RepositoryID,
		RepositoryName:    dump.RepositoryName,
		Indexer:           dump.Indexer,
		IndexerVersion:    dump.IndexerVersion,
		NumParts:          0,
		UploadedParts:     []int{},
		UploadSize:        nil,
		Rank:              nil,
		AssociatedIndexID: dump.AssociatedIndexID,
	}
}

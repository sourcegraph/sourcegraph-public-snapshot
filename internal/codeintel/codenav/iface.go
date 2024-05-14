package codenav

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type UploadService interface {
	GetCompletedUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.CompletedUpload, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
	GetCompletedUploadsByIDs(ctx context.Context, ids []int) (_ []shared.CompletedUpload, err error)
	// The resulting uploads are guaranteed to be unique per (indexer, root) pair,
	// see NOTE(id: closest-uploads-postcondition).
	InferClosestUploads(ctx context.Context, opts shared.UploadMatchingOptions) (_ []shared.CompletedUpload, err error)
}

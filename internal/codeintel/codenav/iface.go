package codenav

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type UploadService interface {
	GetProcessedUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.ProcessedUpload, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
	GetProcessedUploadsByIDs(ctx context.Context, ids []int) (_ []shared.ProcessedUpload, err error)
	InferClosestUploads(ctx context.Context, repositoryID int, commit, path string, exactPath bool, indexer string) (_ []shared.ProcessedUpload, err error)
}

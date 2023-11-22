package summary

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type UploadService interface {
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) ([]uploadsshared.IndexesWithRepositoryNamespace, error)
}

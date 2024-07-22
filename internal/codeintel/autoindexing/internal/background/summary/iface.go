package summary

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type UploadService interface {
	GetRecentUploadsSummary(ctx context.Context, repositoryID int) (upload []shared.UploadsWithRepositoryNamespace, err error)
	GetRecentAutoIndexJobsSummary(ctx context.Context, repositoryID int) ([]uploadsshared.GroupedAutoIndexJobs, error)
}

package graphql

import (
	"context"
	"time"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
)

type UploadService interface {
	sharedresolvers.UploadsService

	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
}

type AutoIndexingService = sharedresolvers.AutoIndexingService
type PolicyService = sharedresolvers.PolicyService

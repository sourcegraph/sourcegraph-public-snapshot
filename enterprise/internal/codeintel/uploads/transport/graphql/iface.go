package graphql

import (
	"context"
	"time"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type UploadService interface {
	sharedresolvers.UploadsService

	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
	DeleteUploadByID(ctx context.Context, id int) (_ bool, err error)
	DeleteUploads(ctx context.Context, opts uploadsshared.DeleteUploadsOptions) (err error)
}

type AutoIndexingService interface {
	sharedresolvers.AutoIndexingService

	GetUnsafeDB() database.DB
}

type PolicyService = sharedresolvers.PolicyService

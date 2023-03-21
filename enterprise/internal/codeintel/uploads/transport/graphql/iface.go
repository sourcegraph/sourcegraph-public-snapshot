package graphql

import (
	"context"
	"time"
)

type UploadService interface {
	GetCommitGraphMetadata(ctx context.Context, repositoryID int) (stale bool, updatedAt *time.Time, err error)
}

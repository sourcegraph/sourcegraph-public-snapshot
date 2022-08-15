package commitgraph

import (
	"context"
	"time"
)

type UploadService interface {
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
	GetDirtyRepositories(ctx context.Context) (map[int]int, error)
	UpdateDirtyRepositories(ctx context.Context, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) (err error)
}

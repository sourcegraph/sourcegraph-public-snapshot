package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/expirer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background/processor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

type UploadService interface {
	GetDirtyRepositories(ctx context.Context) (_ []shared.DirtyRepository, err error)
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
}

type (
	Locker        = commitgraph.Locker
	RepoStore     = processor.RepoStore
	PolicyService = expirer.PolicyService
	PolicyMatcher = expirer.PolicyMatcher
)

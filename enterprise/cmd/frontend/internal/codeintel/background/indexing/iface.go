package indexing

import (
	"context"
	"regexp"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

type DBStore interface {
	IndexableRepositories(ctx context.Context, opts dbstore.IndexableRepositoryQueryOptions) ([]dbstore.IndexableRepository, error)
	GetRepositoriesWithIndexConfiguration(ctx context.Context) ([]int, error)
	RepoUsageStatistics(ctx context.Context) ([]dbstore.RepoUsageStatistics, error)
	ResetIndexableRepositories(ctx context.Context, lastUpdatedBefore time.Time) error
	UpdateIndexableRepository(ctx context.Context, indexableRepository dbstore.UpdateableIndexableRepository, now time.Time) error
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
}

type IndexEnqueuer interface {
	QueueIndex(ctx context.Context, repositoryID int) error
}

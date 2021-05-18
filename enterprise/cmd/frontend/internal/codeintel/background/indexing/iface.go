package indexing

import (
	"context"
	"regexp"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type DBStore interface {
	IndexableRepositories(ctx context.Context, opts dbstore.IndexableRepositoryQueryOptions) ([]dbstore.IndexableRepository, error)
	GetRepositoriesWithIndexConfiguration(ctx context.Context) ([]int, error)
	RepoUsageStatistics(ctx context.Context) ([]dbstore.RepoUsageStatistics, error)
	ResetIndexableRepositories(ctx context.Context, lastUpdatedBefore time.Time) error
	UpdateIndexableRepository(ctx context.Context, indexableRepository dbstore.UpdateableIndexableRepository, now time.Time) error
	GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error)
}

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
}

type gitClient struct {
	client       GitserverClient
	repositoryID int
	commit       string
}

func newGitClient(client GitserverClient, repositoryID int, commit string) gitClient {
	return gitClient{
		client:       client,
		repositoryID: repositoryID,
		commit:       commit,
	}
}

func (s gitClient) ListFiles(ctx context.Context, pattern *regexp.Regexp) ([]string, error) {
	return s.client.ListFiles(ctx, s.repositoryID, s.commit, pattern)
}

func (s gitClient) FileExists(ctx context.Context, file string) (bool, error) {
	return s.client.FileExists(ctx, s.repositoryID, s.commit, file)
}

func (s gitClient) RawContents(ctx context.Context, file string) ([]byte, error) {
	return s.client.RawContents(ctx, s.repositoryID, s.commit, file)
}

type IndexEnqueuer interface {
	QueueIndexesForRepository(ctx context.Context, repositoryID int) error
}

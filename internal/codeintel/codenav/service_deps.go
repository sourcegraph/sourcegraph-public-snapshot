package codenav

import (
	"context"
	"io"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// minimalRepoStore covers the subset of database.RepoStore APIs that we need
// for code navigation.
//
// Prefer calling GetReposSetByIDs instead of calling Get in a loop.
type minimalRepoStore interface {
	Get(context.Context, api.RepoID) (*types.Repo, error)
	GetReposSetByIDs(context.Context, ...api.RepoID) (map[api.RepoID]*types.Repo, error)
}

var _ minimalRepoStore = (database.RepoStore)(nil)

// minimalGitserver covers the subset of gitserver.Client APIs that we
// need for code navigation
type minimalGitserver interface {
	Diff(ctx context.Context, repo api.RepoName, opts gitserver.DiffOptions) (*gitserver.DiffFileIterator, error)
	GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID) (*gitdomain.Commit, error)
	NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (io.ReadCloser, error)
	Stat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error)
}

var _ minimalGitserver = (gitserver.Client)(nil)

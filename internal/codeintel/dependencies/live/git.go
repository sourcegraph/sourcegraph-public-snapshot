package live

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type gitService struct {
	db      database.DB
	checker authz.SubRepoPermissionChecker
}

func NewGitService(db database.DB) dependencies.GitService {
	return &gitService{
		db:      db,
		checker: authz.DefaultSubRepoPermsChecker,
	}
}

func (s *gitService) GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool) ([]*gitdomain.Commit, error) {
	return git.GetCommits(ctx, s.db, repoCommits, ignoreErrors, s.checker)
}

func (s *gitService) LsFiles(ctx context.Context, repo api.RepoName, commits api.CommitID, pathspecs ...gitserver.Pathspec) ([]string, error) {
	return gitserver.NewClient(s.db).LsFiles(ctx, s.checker, repo, commits, pathspecs...)
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	// Note: the sub-repo perms checker is nil here because sub-repo filtering is applied when LsFiles is called
	return git.ArchiveReader(ctx, s.db, nil, repo, opts)
}

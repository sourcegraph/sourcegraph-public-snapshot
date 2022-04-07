package lockfiles

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type GitService interface {
	LsFiles(ctx context.Context, repo api.RepoName, commits api.CommitID, pathspecs ...gitserver.Pathspec) ([]string, error)
	Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error)
}

type gitService struct {
	db      database.DB
	checker authz.SubRepoPermissionChecker
}

func NewDefaultGitService(checker authz.SubRepoPermissionChecker, db database.DB) GitService {
	if checker == nil {
		checker = authz.DefaultSubRepoPermsChecker
	}

	return &gitService{
		db:      db,
		checker: checker,
	}
}

func (s *gitService) LsFiles(ctx context.Context, repo api.RepoName, commits api.CommitID, pathspecs ...gitserver.Pathspec) ([]string, error) {
	return git.LsFiles(ctx, s.db, s.checker, repo, commits, pathspecs...)
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	// Note: the sub-repo perms checker is nil here because sub-repo filtering is applied when LsFiles is called
	return git.ArchiveReader(ctx, s.db, nil, repo, opts)
}

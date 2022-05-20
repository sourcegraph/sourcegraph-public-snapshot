package inference

import (
	"context"
	"io"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type SandboxService interface {
	CreateSandbox(ctx context.Context, opts luasandbox.CreateOptions) (*luasandbox.Sandbox, error)
}

type GitService interface {
	ListFiles(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) ([]string, error)
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

func (s *gitService) ListFiles(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) ([]string, error) {
	return gitserver.NewClient(s.db).ListFiles(ctx, repo, api.CommitID(commit), pattern, authz.DefaultSubRepoPermsChecker)
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	// Note: the sub-repo perms checker is nil here because all paths were already checked via a previous call to s.ListFiles
	return git.ArchiveReader(ctx, s.db, nil, repo, opts)
}

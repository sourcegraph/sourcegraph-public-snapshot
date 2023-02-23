package inference

import (
	"context"
	"io"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
)

type SandboxService interface {
	CreateSandbox(ctx context.Context, opts luasandbox.CreateOptions) (*luasandbox.Sandbox, error)
}

type GitService interface {
	ListFiles(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) ([]string, error)
	Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error)
}

type gitService struct {
	checker authz.SubRepoPermissionChecker
	client  gitserver.Client
}

func NewDefaultGitService(checker authz.SubRepoPermissionChecker) GitService {
	if checker == nil {
		checker = authz.DefaultSubRepoPermsChecker
	}

	return &gitService{
		checker: checker,
		client:  gitserver.NewClient(),
	}
}

func (s *gitService) ListFiles(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) ([]string, error) {
	return s.client.ListFiles(ctx, s.checker, repo, api.CommitID(commit), pattern)
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	// Note: the sub-repo perms checker is nil here because all paths were already checked via a previous call to s.ListFiles
	return s.client.ArchiveReader(ctx, nil, repo, opts)
}

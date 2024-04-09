package inference

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
)

type SandboxService interface {
	CreateSandbox(ctx context.Context, opts luasandbox.CreateOptions) (*luasandbox.Sandbox, error)
}

type GitService interface {
	LsFiles(ctx context.Context, repo api.RepoName, commit string, pathspecs ...gitdomain.Pathspec) ([]string, error)
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
		client:  gitserver.NewClient("codeintel.inference"),
	}
}

func (s *gitService) LsFiles(ctx context.Context, repo api.RepoName, commit string, pathspecs ...gitdomain.Pathspec) ([]string, error) {
	fds, err := s.client.ReadDirPatterns(ctx, repo, api.CommitID(commit), pathspecs)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(fds)/2)
	for _, fd := range fds {
		if fd.IsDir() {
			continue
		}
		files = append(files, fd.Name())
	}
	return files, nil
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	return s.client.ArchiveReader(ctx, repo, opts)
}

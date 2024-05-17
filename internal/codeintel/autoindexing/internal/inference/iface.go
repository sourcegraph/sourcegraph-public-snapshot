package inference

import (
	"context"
	"io"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SandboxService interface {
	CreateSandbox(ctx context.Context, opts luasandbox.CreateOptions) (*luasandbox.Sandbox, error)
}

type GitService interface {
	ResolveRevision(ctx context.Context, repo api.RepoName, spec string) (api.CommitID, error)
	ReadDir(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error)
	Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error)
}

type gitService struct {
	client gitserver.Client
}

func NewDefaultGitService() GitService {
	return &gitService{
		client: gitserver.NewClient("codeintel.inference"),
	}
}

func (s *gitService) ReadDir(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	it, err := s.client.ReadDir(ctx, repo, commit, path, recurse)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	files := make([]fs.FileInfo, 0)
	for {
		file, err := it.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (s *gitService) Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
	return s.client.ArchiveReader(ctx, repo, opts)
}

func (s *gitService) ResolveRevision(ctx context.Context, repo api.RepoName, spec string) (api.CommitID, error) {
	return s.client.ResolveRevision(ctx, repo, spec, gitserver.ResolveRevisionOptions{
		EnsureRevision: false,
	})
}

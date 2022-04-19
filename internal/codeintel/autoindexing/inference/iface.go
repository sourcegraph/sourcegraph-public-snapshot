package inference

import (
	"context"
	"io"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

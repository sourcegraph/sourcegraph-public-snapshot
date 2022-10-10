package ranking

import (
	"context"
	"io"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type GitserverClient interface {
	ListFilesForRepo(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) (_ []string, err error)
	ArchiveReader(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error)
}

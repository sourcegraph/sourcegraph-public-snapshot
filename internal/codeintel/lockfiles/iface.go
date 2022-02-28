package lockfiles

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type GitService interface {
	LsFiles(ctx context.Context, repo api.RepoName, commits api.CommitID, paths ...string) ([]string, error)
	Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error)
}

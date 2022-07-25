package lockfiles

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type GitService interface {
	LsFiles(ctx context.Context, repo api.RepoName, commits api.CommitID, pathspecs ...gitdomain.Pathspec) ([]string, error)
	Archive(ctx context.Context, repo api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error)
}

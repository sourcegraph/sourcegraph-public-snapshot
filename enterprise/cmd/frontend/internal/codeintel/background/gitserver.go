package background

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type gitserverClient interface {
	Head(ctx context.Context, store store.Store, repositoryID int) (string, error)
	ListFiles(ctx context.Context, store store.Store, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, store store.Store, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, store store.Store, repositoryID int, commit, file string) ([]byte, error)
	CommitGraph(ctx context.Context, store store.Store, repositoryID int, options gitserver.CommitGraphOptions) (map[string][]string, error)
}

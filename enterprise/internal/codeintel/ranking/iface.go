package ranking

import (
	"context"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type GitserverClient interface {
	ListFilesForRepo(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) (_ []string, err error)
}

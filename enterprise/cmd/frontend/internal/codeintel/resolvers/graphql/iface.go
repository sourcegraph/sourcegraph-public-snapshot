package graphql

import (
	"context"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
)

type GitserverClient interface {
	policies.GitserverClient
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
}

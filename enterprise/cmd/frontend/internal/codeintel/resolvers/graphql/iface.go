package graphql

import (
	"context"

	"github.com/grafana/regexp"

	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
)

type GitserverClient interface {
	policies.GitserverClient
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
}

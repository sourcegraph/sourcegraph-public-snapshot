package graphql

import (
	"context"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	sharedSymbols "github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type GitserverClient interface {
	policies.GitserverClient
	shared.GitserverClient

	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) ([]*gitdomain.Tag, error)
}

type SymbolService interface {
	References(ctx context.Context, args sharedSymbols.RequestArgs) (_ []sharedSymbols.UploadLocation, _ string, err error)
}

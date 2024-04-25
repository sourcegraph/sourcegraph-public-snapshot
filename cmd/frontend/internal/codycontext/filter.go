package codycontext

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoContentFilter interface {
	// GetMatcher returns a matcher to filter out files that don't satisfy Cody ignore requirements. It also filters
	// down the list of repositories to only those that are allowed to be searched.
	GetMatcher(ctx context.Context, repos []types.RepoIDName) ([]types.RepoIDName, search.CodyFileMatcher, error)
}

func newRepoContentFilter(logger log.Logger, client gitserver.Client) RepoContentFilter {
	if dotcom.SourcegraphDotComMode() {
		return newDotcomFilter(logger, client)
	}
	return newEnterpriseFilter(logger)
}

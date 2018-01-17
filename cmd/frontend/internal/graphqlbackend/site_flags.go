package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

// siteFlagsResolver is embedded in siteResolver. It caches the flag values because they are
// expensive to compute and do not need to be precise or up to date.
type siteFlagsResolver struct{}

func (r *siteResolver) NeedsRepositoryConfiguration(ctx context.Context) (bool, error) {
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	// 🚨 SECURITY: The site alerts may contain sensitive data, so only site
	// admins may view them.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return false, err
	}

	cfg := conf.Get()
	return len(cfg.Github) == 0 && len(cfg.ReposList) == 0 && cfg.GitoliteHosts == "" && cfg.GitOriginMap == "", nil
}

func (r *siteResolver) NoRepositoriesEnabled(ctx context.Context) (bool, error) {
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	// 🚨 SECURITY: The site alerts may contain sensitive data, so only site
	// admins may view them.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return false, err
	}

	// Fastest way to see if even a single enabled repository exists.
	repos, err := db.Repos.List(ctx, &db.ReposListOptions{
		Enabled:     true,
		Disabled:    false,
		ListOptions: sourcegraph.ListOptions{PerPage: 1},
	})
	if err != nil {
		return false, err
	}
	return len(repos) == 0, nil
}

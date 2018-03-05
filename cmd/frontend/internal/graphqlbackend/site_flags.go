package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
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

	return needsRepositoryConfiguration(), nil
}

func needsRepositoryConfiguration() bool {
	cfg := conf.Get()
	return len(cfg.Github) == 0 && len(cfg.Gitlab) == 0 && len(cfg.ReposList) == 0 && len(cfg.AwsCodeCommit) == 0 && cfg.GitoliteHosts == "" && cfg.GitOriginMap == ""
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

	return noRepositoriesEnabled(ctx)
}

func noRepositoriesEnabled(ctx context.Context) (bool, error) {
	// Fastest way to see if even a single enabled repository exists.
	repos, err := db.Repos.List(ctx, db.ReposListOptions{
		Enabled:     true,
		Disabled:    false,
		LimitOffset: &db.LimitOffset{Limit: 1},
	})
	if err != nil {
		return false, err
	}
	return len(repos) == 0, nil
}

var isExternalAuthEnabled = func() bool {
	oidcProvider := conf.AuthOpenIDConnect()
	if oidcProvider != nil && oidcProvider.ClientID != "" && oidcProvider.ClientSecret != "" {
		return true
	}

	samlProvider := conf.AuthSAML()
	if samlProvider != nil && samlProvider.IdentityProviderMetadataURL != "" && samlProvider.ServiceProviderCertificate == "" && samlProvider.ServiceProviderPrivateKey == "" {
		return true
	}

	return false
}()

func (*siteResolver) ExternalAuthEnabled() bool { return isExternalAuthEnabled }

func (*siteResolver) DisableBuiltInSearches() bool {
	return conf.Get().DisableExampleSearches || conf.Get().DisableBuiltInSearches
}

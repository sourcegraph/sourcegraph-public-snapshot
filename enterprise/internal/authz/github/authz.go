package github

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ExternalConnection is a composite object of a GITHUB kind external service and
// parsed connection information.
type ExternalConnection struct {
	*types.ExternalService
	*types.GitHubConnection
}

// NewAuthzProviders returns the set of GitHub authz providers derived from the connections.
//
// It also returns any simple validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
//
// This constructor does not and should not directly check connectivity to external services - if
// desired, callers should use `(*Provider).ValidateConnection` directly to get warnings related
// to connection issues.
func NewAuthzProviders(
	externalServicesStore database.ExternalServiceStore,
	conns []*ExternalConnection,
	authProviders []schema.AuthProviders,
	enableGithubInternalRepoVisibility bool,
) (ps []authz.Provider, problems []string, warnings []string) {
	// Auth providers (i.e. login mechanisms)
	githubAuthProviders := make(map[string]*schema.GitHubAuthProvider)
	for _, p := range authProviders {
		if p.Github != nil {
			var id string
			ghURL, err := url.Parse(p.Github.GetURL())
			if err != nil {
				// error reporting for this should happen elsewhere, for now just use what is given
				id = p.Github.GetURL()
			} else {
				// use codehost normalized URL as ID
				ch := extsvc.NewCodeHost(ghURL, p.Github.Type)
				id = ch.ServiceID
			}
			githubAuthProviders[id] = p.Github
		}
	}

	for _, c := range conns {
		// Initialize authz (permissions) provider.
		p, err := newAuthzProvider(externalServicesStore, c)
		if err != nil {
			problems = append(problems, err.Error())
		}
		if p == nil {
			continue
		}

		// We want to make the feature flag available to the GitHub provider, but at the same time
		// also not use the global conf.SiteConfig which is discouraged and could cause race
		// conditions. As a result, we use a temporary hack by setting this on the provider for now.
		p.enableGithubInternalRepoVisibility = enableGithubInternalRepoVisibility

		// Permissions require a corresponding GitHub OAuth provider. Without one, repos
		// with restricted permissions will not be visible to non-admins.
		if authProvider, exists := githubAuthProviders[p.ServiceID()]; !exists {
			warnings = append(warnings,
				fmt.Sprintf("GitHub config for %[1]s has `authorization` enabled, "+
					"but no authentication provider matching %[1]q was found. "+
					"Check the [**site configuration**](/site-admin/configuration) to "+
					"verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for %[1]s.",
					p.ServiceID()))
		} else if p.groupsCache != nil && !authProvider.AllowGroupsPermissionsSync {
			// Groups permissions requires auth provider to request the correct scopes.
			warnings = append(warnings,
				fmt.Sprintf("GitHub config for %[1]s has `authorization.groupsCacheTTL` enabled, but "+
					"the authentication provider matching %[1]q does not have `allowGroupsPermissionsSync` enabled. "+
					"Update the [**site configuration**](/site-admin/configuration) in the appropriate entry "+
					"in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) to enable this.",
					p.ServiceID()))
			// Forcibly disable groups cache.
			p.groupsCache = nil
		}

		// Register this provider.
		ps = append(ps, p)
	}

	return ps, problems, warnings
}

// newAuthzProvider instantiates a provider, or returns nil if authorization is disabled.
// Errors returned are "serious problems".
func newAuthzProvider(
	externalServicesStore database.ExternalServiceStore,
	c *ExternalConnection,
) (*Provider, error) {
	if c.Authorization == nil {
		return nil, nil
	}

	baseURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, errors.Errorf("could not parse URL for GitHub instance %q: %s", c.Url, err)
	}

	if c.GithubAppInstallationID != "" {
		installationID, err := strconv.ParseInt(c.GithubAppInstallationID, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse installation ID")
		}

		gitHubAppConfig := conf.SiteConfig().GitHubApp
		if repos.IsGitHubAppEnabled(gitHubAppConfig) {
			return newAppProvider(externalServicesStore, c.ExternalService, c.GitHubConnection.URN, baseURL, gitHubAppConfig.AppID, gitHubAppConfig.PrivateKey, installationID, nil)
		}

		return nil, errors.Errorf("connection contains an GitHub App installation ID while GitHub App for Sourcegraph is not enabled")
	}

	// Disable by default for now
	if c.Authorization.GroupsCacheTTL <= 0 {
		c.Authorization.GroupsCacheTTL = -1
	}

	ttl := time.Duration(c.Authorization.GroupsCacheTTL) * time.Hour
	return NewProvider(c.GitHubConnection.URN, ProviderOptions{
		GitHubURL:      baseURL,
		BaseToken:      c.Token,
		GroupsCacheTTL: ttl,
	}), nil
}

// ValidateAuthz validates the authorization fields of the given GitHub external
// service config.
func ValidateAuthz(c *types.GitHubConnection) error {
	_, err := newAuthzProvider(nil, &ExternalConnection{GitHubConnection: c})
	return err
}

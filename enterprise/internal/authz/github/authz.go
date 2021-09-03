package github

import (
	"fmt"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of GitHub authz providers derived from the connections.
// It also returns any validation problems with the config, separating these into "serious problems" and
// "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
func NewAuthzProviders(
	conns []*types.GitHubConnection,
	authProviders []schema.AuthProviders,
) (ps []authz.Provider, problems []string, warnings []string) {
	// Permissions providers
	for _, c := range conns {
		p, err := newAuthzProvider(c.URN, c.Authorization, c.Url, c.Token)
		if err != nil {
			problems = append(problems, err.Error())
		} else if p != nil {
			ps = append(ps, p)
		}
	}

	// Auth providers (i.e. login mechanisms)
	githubAuthProviders := make(map[string]struct{})
	for _, p := range authProviders {
		if p.Github != nil {
			var id string
			ghURL, err := url.Parse(p.Github.Url)
			if err != nil {
				// error reporting for this should happen elsewhere, for now just use what is given
				id = p.Github.Url
			} else {
				// use codehost normalized URL as ID
				ch := extsvc.NewCodeHost(ghURL, p.Github.Type)
				id = ch.ServiceID
			}
			githubAuthProviders[id] = struct{}{}
		}
	}

	for _, p := range ps {
		// Permissions require a corresponding GitHub OAuth provider. Without one, repos
		// with restricted permissions will not be visible to non-admins.
		if _, exists := githubAuthProviders[p.ServiceID()]; !exists {
			warnings = append(warnings,
				fmt.Sprintf("Did not find authentication provider matching %[1]q. "+
					"Check the [**site configuration**](/site-admin/configuration) to "+
					"verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for %[1]s.",
					p.ServiceID()))
		}

		// Check for other validation issues.
		for _, problem := range p.Validate() {
			warnings = append(warnings, fmt.Sprintf("GitHub config for %s was invalid: %s", p.ServiceID(), problem))
		}
	}

	return ps, problems, warnings
}

// newAuthzProvider instantiates a provider, or returns nil if authorization is disabled.
// Errors returned are "serious problems".
func newAuthzProvider(urn string, a *schema.GitHubAuthorization, instanceURL, token string) (authz.Provider, error) {
	if a == nil {
		return nil, nil
	}

	ghURL, err := url.Parse(instanceURL)
	if err != nil {
		return nil, errors.Errorf("Could not parse URL for GitHub instance %q: %s", instanceURL, err)
	}

	// Disable by default for now
	if a.GroupsCacheTTL <= 0 {
		a.GroupsCacheTTL = -1
	}

	ttl := time.Duration(a.GroupsCacheTTL) * time.Hour

	return NewProvider(urn, ProviderOptions{
		GitHubURL:      ghURL,
		BaseToken:      token,
		GroupsCacheTTL: ttl,
	}), nil
}

// ValidateAuthz validates the authorization fields of the given GitHub external
// service config.
func ValidateAuthz(cfg *schema.GitHubConnection) error {
	_, err := newAuthzProvider("", cfg.Authorization, cfg.Url, cfg.Token)
	return err
}

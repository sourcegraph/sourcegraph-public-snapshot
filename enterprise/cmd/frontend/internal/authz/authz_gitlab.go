package authz

import (
	"fmt"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	permgl "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func gitlabProvidersFromConfig(cfg *conf.Unified) (
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	// Authorization (i.e., permissions) providers
	for _, gl := range cfg.Gitlab {
		if gl.Authorization == nil {
			continue
		}

		glURL, err := url.Parse(gl.Url)
		if err != nil {
			seriousProblems = append(seriousProblems, fmt.Sprintf("Could not parse URL for GitLab instance %q: %s", gl.Url, err))
			continue // omit authz provider if could not parse URL
		}

		var ttl time.Duration
		ttl, warnings = parseTTLOrDefault(gl.Authorization.Ttl, 3*time.Hour, warnings)

		op := permgl.GitLabAuthzProviderOp{
			BaseURL:   glURL,
			SudoToken: gl.Token,
			AuthnConfigID: auth.ProviderConfigID{
				ID:   gl.Authorization.AuthnProvider.ConfigID,
				Type: gl.Authorization.AuthnProvider.Type,
			},
			GitLabProvider: gl.Authorization.AuthnProvider.GitlabProvider,
			CacheTTL:       ttl,
			MockCache:      nil,
		}
		if gl.Authorization.AuthnProvider.ConfigID == "" {
			// Note: In the future when we support sign-in via GitLab, we can check if that is
			// enabled and instead fall back to that.
			if env.InsecureDev {
				log15.Warn("Using username matching for debugging purposes, because `authz.authnProvider.configID` in the config was empty. This should ONLY occur in a development build.")
				op.UseNativeUsername = true
			} else {
				seriousProblems = append(seriousProblems, "`authz.authnProvider.configID` was empty. No users will be granted access to these repositories.")
			}
		} else if gl.Authorization.AuthnProvider.Type == "" {
			seriousProblems = append(seriousProblems, "`authz.authnProvider.type` was not specified, which means GitLab users cannot be resolved.")
		} else if gl.Authorization.AuthnProvider.GitlabProvider == "" {
			seriousProblems = append(seriousProblems, "`authz.authnProvider.gitlabProvider` was not specified, which means GitLab users cannot be resolved.")
		} else {
			// Best-effort determine if the authz.authnConfigID field refers to an item in auth.provider
			found := false
			for _, p := range cfg.Critical.AuthProviders {
				if p.Openidconnect != nil && p.Openidconnect.ConfigID == gl.Authorization.AuthnProvider.ConfigID && p.Openidconnect.Type == gl.Authorization.AuthnProvider.Type {
					found = true
					break
				}
				if p.Saml != nil && p.Saml.ConfigID == gl.Authorization.AuthnProvider.ConfigID && p.Saml.Type == gl.Authorization.AuthnProvider.Type {
					found = true
					break
				}
			}
			if !found {
				seriousProblems = append(seriousProblems, fmt.Sprintf("Could not find item in `auth.providers` with config ID %q and type %q", gl.Authorization.AuthnProvider.ConfigID, gl.Authorization.AuthnProvider.Type))
			}
		}

		authzProviders = append(authzProviders, NewGitLabProvider(op))
	}
	for _, provider := range authzProviders {
		for _, problem := range provider.Validate() {
			warnings = append(warnings, fmt.Sprintf("GitLab config for %s was invalid: %s", provider.ServiceID(), problem))
		}
	}
	return authzProviders, seriousProblems, warnings
}

// NewGitLabProvider is a mockable constructor for new GitLabAuthzProvider instances.
var NewGitLabProvider = func(op permgl.GitLabAuthzProviderOp) authz.Provider {
	return permgl.NewProvider(op)
}

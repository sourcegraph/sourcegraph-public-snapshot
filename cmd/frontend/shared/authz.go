package shared

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
	permgl "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	conf.ContributeValidator(func(cfg schema.SiteConfiguration) []string {
		_, _, seriousProblems, warnings := providersFromConfig(&cfg)
		return append(seriousProblems, warnings...)
	})
	conf.Watch(func() {
		allowAccessByDefault, authzProviders, _, _ := providersFromConfig(conf.Get())
		authz.SetProviders(allowAccessByDefault, authzProviders)
	})
}

// providersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings".  "Serious problems" are those that should make Sourcegraph set
// authz.allowAccessByDefault to false. "Warnings" are all other validation problems.
func providersFromConfig(cfg *schema.SiteConfiguration) (
	allowAccessByDefault bool,
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	allowAccessByDefault = true
	defer func() {
		if len(seriousProblems) > 0 {
			log15.Error("Repository authz config was invalid (errors are visible in the UI as an admin user, you should fix ASAP). Restricting access to repositories by default for now to be safe.")
			allowAccessByDefault = false
		}
	}()

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
		if innerMatcher := strings.TrimSuffix(strings.TrimPrefix(gl.Authorization.Matcher, "*/"), "/*"); strings.Contains(innerMatcher, "*") {
			seriousProblems = append(seriousProblems, fmt.Sprintf("GitLab connection %q `permission.matcher` includes an interior wildcard \"*\", which will be interpreted as a string literal, rather than a pattern matcher. Only the prefix \"*/\" or the suffix \"/*\" is supported for pattern matching.", gl.Url))
		}

		var ttl time.Duration
		if gl.Authorization.Ttl == "" {
			ttl = time.Hour * 3
		} else {
			ttl, err = time.ParseDuration(gl.Authorization.Ttl)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Could not parse time duration %q, falling back to 3 hours.", gl.Authorization.Ttl))
				ttl = time.Hour * 3
			}
		}

		op := permgl.GitLabAuthzProviderOp{
			BaseURL:   glURL,
			SudoToken: gl.Token,
			AuthnConfigID: auth.ProviderConfigID{
				ID:   gl.Authorization.AuthnProvider.ConfigID,
				Type: gl.Authorization.AuthnProvider.Type,
			},
			GitLabProvider: gl.Authorization.AuthnProvider.GitlabProvider,
			MatchPattern:   gl.Authorization.Matcher,
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
			for _, p := range cfg.AuthProviders {
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

		authzProviders = append(authzProviders, NewProvider(op))
	}

	return allowAccessByDefault, authzProviders, seriousProblems, warnings
}

// NewProvider is a mockable constructor for new GitLabAuthzProvider instances.
var NewProvider = func(op permgl.GitLabAuthzProviderOp) authz.Provider {
	return permgl.NewProvider(op)
}

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
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	conf.ContributeValidator(func(cfg schema.SiteConfiguration) []string {
		_, _, seriousProblems, warnings := providersFromConfig(&cfg)
		return append(seriousProblems, warnings...)
	})
	conf.Watch(func() {
		permissionsAllowByDefault, authzProviders, _, _ := providersFromConfig(conf.Get())
		authz.SetProviders(permissionsAllowByDefault, authzProviders)
	})
}

// providersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings".  "Serious problems" are those that should make Sourcegraph set
// authz.permissionsAllowByDefault to false. "Warnings" are all other validation problems.
func providersFromConfig(cfg *schema.SiteConfiguration) (
	permissionsAllowByDefault bool,
	authzProviders []authz.AuthzProvider,
	seriousProblems []string,
	warnings []string,
) {
	permissionsAllowByDefault = true
	defer func() {
		if len(seriousProblems) > 0 {
			log15.Error("Repository authz config was invalid (errors are visible in the UI as an admin user, you should fix ASAP). Restricting access to repositories by default for now to be safe.")
			permissionsAllowByDefault = false
		}
	}()

	// Authorization (i.e., permissions) providers
	for _, gl := range cfg.Gitlab {
		if gl.Authz == nil {
			continue
		}

		glURL, err := url.Parse(gl.Url)
		if err != nil {
			seriousProblems = append(seriousProblems, fmt.Sprintf("Could not parse URL for GitLab instance %q: %s", gl.Url, err))
			continue // omit authz provider if could not parse URL
		}
		if innerMatcher := strings.TrimSuffix(strings.TrimPrefix(gl.Authz.Matcher, "*/"), "/*"); strings.Contains(innerMatcher, "*") {
			seriousProblems = append(seriousProblems, fmt.Sprintf("GitLab connection %q `permission.matcher` includes an interior wildcard \"*\", which will be interpreted as a string literal, rather than a pattern matcher. Only the prefix \"*/\" or the suffix \"/*\" is supported for pattern matching.", gl.Url))
		}

		var ttl time.Duration
		if gl.Authz.Ttl == "" {
			ttl = time.Hour * 3
		} else {
			ttl, err = time.ParseDuration(gl.Authz.Ttl)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Could not parse time duration %q, falling back to 3 hours.", gl.Authz.Ttl))
				ttl = time.Hour * 3
			}
		}

		op := permgl.GitLabAuthzProviderOp{
			BaseURL:   glURL,
			SudoToken: gl.Token,
			AuthnConfigID: auth.ProviderConfigID{
				ID:   gl.Authz.AuthnProvider.ConfigID,
				Type: gl.Authz.AuthnProvider.Type,
			},
			GitLabProvider: gl.Authz.AuthnProvider.GitlabProvider,
			MatchPattern:   gl.Authz.Matcher,
			CacheTTL:       ttl,
			MockCache:      nil,
		}
		if gl.Authz.AuthnProvider.ConfigID == "" {
			// If no authn provider is specified, we fall back to insecure username matching, which
			// should only be used for testing purposes. In the future when we support sign-in via
			// GitLab, we can check if that is enabled and instead fall back to that.
			seriousProblems = append(seriousProblems, "Security issue: `authz.authnProvider.configID` was empty. Falling back to using username equality for permissions, which is insecure.")
			op.UseNativeUsername = true
		} else if gl.Authz.AuthnProvider.Type == "" {
			seriousProblems = append(seriousProblems, "`authz.authnProvider.type` was not specified, which means GitLab users cannot be resolved.")
		} else if gl.Authz.AuthnProvider.GitlabProvider == "" {
			seriousProblems = append(seriousProblems, "`authz.authnProvider.gitlabProvider` was not specified, which means GitLab users cannot be resolved.")
		} else {
			// Best-effort determine if the authz.authnConfigID field refers to an item in auth.provider
			found := false
			for _, p := range cfg.AuthProviders {
				if p.Openidconnect != nil && p.Openidconnect.ConfigID == gl.Authz.AuthnProvider.ConfigID && p.Openidconnect.Type == gl.Authz.AuthnProvider.Type {
					found = true
					break
				}
				if p.Saml != nil && p.Saml.ConfigID == gl.Authz.AuthnProvider.ConfigID && p.Saml.Type == gl.Authz.AuthnProvider.Type {
					found = true
					break
				}
			}
			if !found {
				seriousProblems = append(seriousProblems, fmt.Sprintf("Could not find item in `auth.providers` with config ID %q and type %q", gl.Authz.AuthnProvider.ConfigID, gl.Authz.AuthnProvider.Type))
			}
		}

		authzProviders = append(authzProviders, NewGitLabAuthzProvider(op))
	}

	return permissionsAllowByDefault, authzProviders, seriousProblems, warnings
}

// NewGitLabAuthzProvider is a mockable constructor for new GitLabAuthzProvider instances.
var NewGitLabAuthzProvider = func(op permgl.GitLabAuthzProviderOp) authz.AuthzProvider {
	return permgl.NewGitLabAuthzProvider(op)
}

package confvalidators

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/azureoauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/bitbucketcloudoauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/gerrit"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/githuboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/gitlaboauth"
	"github.com/sourcegraph/sourcegraph/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

func validateAuthProviders(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if len(c.SiteConfig().AuthProviders) == 0 {
		problems = append(problems, conf.NewSiteProblem("no auth providers set (all access will be forbidden)"))
	}

	// Validate that `auth.enableUsernameChanges` is not set if SSO is configured
	if conf.HasExternalAuthProvider(c) && c.SiteConfig().AuthEnableUsernameChanges {
		problems = append(problems, conf.NewSiteProblem("`auth.enableUsernameChanges` must not be true if external auth providers are set in `auth.providers`"))
	}

	return problems
}

func validateHttpHeaderAuth(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	var httpHeaderAuthProviders int
	for _, p := range c.SiteConfig().AuthProviders {
		if p.HttpHeader != nil {
			httpHeaderAuthProviders++
		}
	}
	if httpHeaderAuthProviders >= 2 {
		problems = append(problems, conf.NewSiteProblem(`at most 1 HTTP header auth provider may be set in site config`))
	}
	return problems
}

func validateOIDCConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	var loggedNeedsExternalURL bool
	for _, p := range c.SiteConfig().AuthProviders {
		if p.Openidconnect != nil && c.SiteConfig().ExternalURL == "" && !loggedNeedsExternalURL {
			problems = append(problems, conf.NewSiteProblem("openidconnect auth provider requires `externalURL` to be set to the external URL of your site (example: https://sourcegraph.example.com)"))
			loggedNeedsExternalURL = true
		}
	}

	seen := map[schema.OpenIDConnectAuthProvider]int{}
	for i, p := range c.SiteConfig().AuthProviders {
		if p.Openidconnect != nil {
			if j, ok := seen[*p.Openidconnect]; ok {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("OpenID Connect auth provider at index %d is duplicate of index %d, ignoring", i, j)))
			} else {
				seen[*p.Openidconnect] = i
			}
		}
	}

	return problems
}

func validateUserPasswdAuth(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	var builtinAuthProviders int
	for _, p := range c.SiteConfig().AuthProviders {
		if p.Builtin != nil {
			builtinAuthProviders++
		}
	}
	if builtinAuthProviders >= 2 {
		problems = append(problems, conf.NewSiteProblem(`at most 1 builtin auth provider may be used`))
	}
	return problems
}

func validateSourcegraphOperatorAuth(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	cloudSiteConfig := cloud.SiteConfig()
	if !cloudSiteConfig.SourcegraphOperatorAuthProviderEnabled() {
		return
	}

	if c.SiteConfig().ExternalURL == "" {
		problems = append(
			problems,
			conf.NewSiteProblem("Sourcegraph Operator authentication provider requires `externalURL` to be set to the external URL of your site (example: https://sourcegraph.example.com)"),
		)
	}
	return problems
}

func validateSAMLAuth(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	var loggedNeedsExternalURL bool
	for _, p := range c.SiteConfig().AuthProviders {
		if p.Saml != nil && c.SiteConfig().ExternalURL == "" && !loggedNeedsExternalURL {
			problems = append(problems, conf.NewSiteProblem("saml auth provider requires `externalURL` to be set to the external URL of your site (example: https://sourcegraph.example.com)"))
			loggedNeedsExternalURL = true
		}
	}

	seen := map[string]int{}
	for i, p := range c.SiteConfig().AuthProviders {
		if p.Saml != nil {
			// we can ignore errors: converting to JSON must work, as we parsed from JSON before
			bytes, _ := json.Marshal(*p.Saml)
			key := string(bytes)
			if j, ok := seen[key]; ok {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("SAML auth provider at index %d is duplicate of index %d, ignoring", i, j)))
			} else {
				seen[key] = i
			}
		}
	}

	return problems
}

func validateAzureOAuthProvider(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	_, problems = azureoauth.ParseConfig(logger, cfg, db)
	return problems
}

func validateGitLabOAuthProvider(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	_, problems = gitlaboauth.ParseConfig(logger, cfg, db)
	return problems
}

func validateGitHubOAuthProvider(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	_, problems = githuboauth.ParseConfig(logger, cfg, db)
	return problems
}

func validateGerritAuthProvider(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	_, problems = gerrit.ParseConfig(logger, cfg, db)
	return problems
}

func validateBitbucketCloudOAuthProvider(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	_, problems = bitbucketcloudoauth.ParseConfig(logger, cfg, db)
	return problems
}

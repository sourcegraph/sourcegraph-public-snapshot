package sourcegraphoperator

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// GetOIDCProvider looks up the registered Sourcegraph Operator authentication
// provider with the given ID and returns the underlying *openidconnect.Provider.
// It returns nil if no such provider exists.
func GetOIDCProvider(id string) *openidconnect.Provider {
	p, ok := providers.GetProviderByConfigID(
		providers.ConfigID{
			Type: auth.SourcegraphOperatorProviderType,
			ID:   id,
		},
	).(*provider)
	if ok {
		return p.Provider
	}
	return nil
}

// Init registers Sourcegraph Operator handlers and providers.
func Init() {
	cloudSiteConfig := cloud.SiteConfig()
	if !cloudSiteConfig.SourcegraphOperatorAuthProviderEnabled() {
		return
	}

	conf.ContributeValidator(validateConfig)

	p := NewProvider(*cloudSiteConfig.AuthProviders.SourcegraphOperator, httpcli.ExternalClient)
	providers.Update(auth.SourcegraphOperatorProviderType, []providers.Provider{p})
}

func validateConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if c.SiteConfig().ExternalURL == "" {
		problems = append(
			problems,
			conf.NewSiteProblem("Sourcegraph Operator authentication provider requires `externalURL` to be set to the external URL of your site (example: https://sourcegraph.example.com)"),
		)
	}
	return problems
}

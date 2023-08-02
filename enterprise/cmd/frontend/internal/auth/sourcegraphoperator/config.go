package sourcegraphoperator

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/openidconnect"
	osssourcegraphoperator "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/cloud"
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

	p := NewProvider(*cloudSiteConfig.AuthProviders.SourcegraphOperator)
	logger := log.Scoped(auth.SourcegraphOperatorProviderType, "Sourcegraph Operator config watch")
	go func() {
		if err := p.Refresh(context.Background()); err != nil {
			logger.Error("failed to fetch Sourcegraph Operator service provider metadata", log.Error(err))
		}
	}()
	providers.Update(auth.SourcegraphOperatorProviderType, []providers.Provider{p})

	// Register enterprise handler implementation in OSS
	osssourcegraphoperator.RegisterAddSourcegraphOperatorExternalAccountHandler(addSourcegraphOperatorExternalAccount)
}

package sourcegraphoperator

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ProviderType is the unique identifier of the Sourcegraph Operator
// authentication provider.
const ProviderType = "sourcegraph-operator"

// provider is an implementation of providers.Provider for the Sourcegraph
// Operator authentication.
type provider struct {
	config schema.SourcegraphOperatorAuthProvider
	*openidconnect.Provider
}

// NewProvider creates and returns a new Sourcegraph Operator authentication
// provider using the given config.
func NewProvider(config schema.SourcegraphOperatorAuthProvider) providers.Provider {
	allowSignUp := true
	return &provider{
		config: config,
		Provider: openidconnect.NewProvider(
			schema.OpenIDConnectAuthProvider{
				AllowSignup:        &allowSignUp,
				ClientID:           config.ClientID,
				ClientSecret:       config.ClientSecret,
				ConfigID:           ProviderType,
				DisplayName:        "Sourcegraph Operators",
				Issuer:             config.Issuer,
				RequireEmailDomain: "sourcegraph.com",
				Type:               ProviderType,
			},
			authPrefix,
		).(*openidconnect.Provider),
	}
}

// Config implements providers.Provider.
func (p *provider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		SourcegraphOperator: &p.config,
	}
}

// LifecycleDuration returns the converted lifecycle duration from given minutes.
// It returns the default duration (60 minutes) if the given minutes is
// non-positive.
func LifecycleDuration(minutes int) time.Duration {
	if minutes <= 0 {
		return 60 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

func (p *provider) lifecycleDuration() time.Duration {
	return LifecycleDuration(p.config.LifecycleDuration)
}

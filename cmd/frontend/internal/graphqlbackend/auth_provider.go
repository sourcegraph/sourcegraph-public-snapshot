package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// authProviderResolver resolves an auth provider.
//
// 🚨 SECURITY: The auth provider's value contains secrets that are only visible to site admins and
// MUST NOT be shown to other users.
type authProviderResolver struct {
	authProvider schema.AuthProviders
}

func (r *authProviderResolver) DisplayName() string {
	// 🚨 SECURITY: This method MUST NOT return any secret data from the auth provider configuration
	// values (or if it does, an is-site-admin check must be added).
	switch {
	case r.authProvider.Builtin != nil:
		return "Builtin username-password authentication"
	case r.authProvider.Openidconnect != nil:
		return "OpenID Connect authentication"
	case r.authProvider.Saml != nil:
		return "SAML authentication"
	case r.authProvider.HttpHeader != nil:
		return "Web authentication proxy"
	default:
		return "Unknown"
	}
}

func (r *authProviderResolver) ServiceType() string {
	// 🚨 SECURITY: This method MUST NOT return any secret data from the auth provider configuration
	// values (or if it does, an is-site-admin check must be added).
	return conf.AuthProviderType(r.authProvider)
}

func (r *authProviderResolver) ServiceID() string {
	// 🚨 SECURITY: This method MUST NOT return any secret data from the auth provider configuration
	// values (or if it does, an is-site-admin check must be added).
	switch {
	case r.authProvider.Builtin != nil:
		return ""
	case r.authProvider.Openidconnect != nil:
		return r.authProvider.Openidconnect.Issuer
	case r.authProvider.Saml != nil:
		return r.authProvider.Saml.IdentityProviderMetadataURL
	case r.authProvider.HttpHeader != nil:
		return "http-header"
	default:
		return ""
	}
}

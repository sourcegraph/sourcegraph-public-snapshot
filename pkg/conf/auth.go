package conf

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

// AuthProviderType returns the type string for the auth provider.
func AuthProviderType(p schema.AuthProviders) string {
	switch {
	case p.Builtin != nil:
		return p.Builtin.Type
	case p.Openidconnect != nil:
		return p.Openidconnect.Type
	case p.Saml != nil:
		return p.Saml.Type
	case p.HttpHeader != nil:
		return p.HttpHeader.Type
	default:
		return ""
	}
}

// AuthProviders returns the configured auth providers. It supports auth.providers (highest
// precedence), auth.provider (backcompat, middle precedence), and legacy saml*/oidc* properties
// (lowest precedence).
//
// 🚨 SECURITY: If len(AuthProviders()) == 0, then no auth provider is set. In this case, all access
// MUST be forbidden. The goal is to prevent a config typo from resulting in a data breach.
func AuthProviders() []schema.AuthProviders { return AuthProvidersFromConfig(Get()) }

// AuthProvidersFromConfig is like AuthProvider, except it accepts a site configuration input value
// instead of using the current global value.
func AuthProvidersFromConfig(c *schema.SiteConfiguration) []schema.AuthProviders {
	if c.AuthProviders != nil {
		if !MultipleAuthProvidersEnabledFromConfig(c) && len(c.AuthProviders) >= 1 {
			// Only return first auth provider because the multipleAuthProviders experiment is
			// disabled.
			return c.AuthProviders[:1]
		}
		return c.AuthProviders
	}

	// BACKCOMPAT: A singular provider was configured.
	p := authProvider(c)
	if p == (schema.AuthProviders{}) || AuthProviderType(p) == "" {
		return nil // no auth providers were set
	}
	return []schema.AuthProviders{p}
}

// authProvider (see AuthProvider).
func authProvider(c *schema.SiteConfiguration) schema.AuthProviders {
	switch {
	case len(c.AuthProviders) > 0:
		return c.AuthProviders[0]
	case c.AuthProvider != "":
		return authProviderSingular(c)
	default:
		return schema.AuthProviders{} // none set (this forbids all access)
	}
}

// authProviderSingular returns an auth provider config specified using the deprecated singleton
// "auth.provider" property and other "auth.{openIDConnect,saml,userIdentityHTTPHeader,allowSignup}"
// properties.
func authProviderSingular(c *schema.SiteConfiguration) schema.AuthProviders {
	switch c.AuthProvider {
	case "openidconnect":
		var o schema.OpenIDConnectAuthProvider
		if c.AuthOpenIDConnect != nil {
			o = *c.AuthOpenIDConnect
		}
		o.Type = "openidconnect"
		return schema.AuthProviders{Openidconnect: &o}
	case "saml":
		var o schema.SAMLAuthProvider
		if c.AuthSaml != nil {
			o = *c.AuthSaml
		}
		o.Type = "saml"
		return schema.AuthProviders{Saml: &o}
	case "http-header":
		return schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{
			Type:           "http-header",
			UsernameHeader: c.AuthUserIdentityHTTPHeader,
		}}
	case "builtin":
		return schema.AuthProviders{Builtin: &schema.BuiltinAuthProvider{
			Type:        "builtin",
			AllowSignup: c.AuthAllowSignup,
		}}
	default:
		// 🚨 SECURITY: This means "forbid all access". See the func AuthProviders SECURITY note.
		return schema.AuthProviders{}
	}
}

// AuthPublic reports whether the site is public. Currently only the builtin auth provider allows
// sites to be public. AuthPublic only returns true if auth.public (in site config) is true *and*
// there is a builtin auth provider.
func AuthPublic() bool { return authPublic(Get()) }
func authPublic(c *schema.SiteConfiguration) bool {
	for _, p := range AuthProvidersFromConfig(c) {
		if p.Builtin != nil && c.AuthPublic {
			return true
		}
	}
	return false
}

// AuthAllowSignup reports whether the site allows signup. Currently only the builtin auth
// provider allows signup. AuthAllowSignup returns true if auth.allowSignup is true OR if
// auth.providers' builtin provider has allowSignup true (in site config).
func AuthAllowSignup() bool { return authAllowSignup(Get()) }
func authAllowSignup(c *schema.SiteConfiguration) bool {
	for _, p := range AuthProvidersFromConfig(c) {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return false
}

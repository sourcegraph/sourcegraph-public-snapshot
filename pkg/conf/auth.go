package conf

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

// AuthProviderType returns the type string for the auth provider.
func AuthProviderType() string {
	p := AuthProvider()
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

// AuthProvider returns the auth provider config. It supports auth.providers (highest precedence),
// auth.provider (backcompat, middle precedence), and legacy saml*/oidc* properties (lowest
// precedence).
//
// A nil return value means "forbid all access to server" NOT "server is 100% public, no auth
// needed".
//
// 🚨 SECURITY: The return value schema.AuthProviders{} indicates that no auth provider is set. In
// this case, all access MUST be forbidden. The goal is to prevent a config typo from resulting in a
// data breach.
func AuthProvider() schema.AuthProviders {
	return authProvider(Get())
}

// authProvider (see AuthProvider).
func authProvider(c *schema.SiteConfiguration) schema.AuthProviders {
	switch {
	case len(c.AuthProviders) > 0:
		return c.AuthProviders[0]
	case c.AuthProvider != "":
		return authProviderSingular(c)
	default:
		return authProviderLegacy(c)
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
		// 🚨 SECURITY: This is means "forbid all access". See the func authProvider SECURITY note.
		return schema.AuthProviders{}
	}
}

// authProviderLegacy returns an auth provider config from the deprecated, legacy `oidc*` and
// `saml*` properties.
func authProviderLegacy(c *schema.SiteConfiguration) schema.AuthProviders {
	if c.OidcClientID != "" || c.OidcClientSecret != "" || c.OidcProvider != "" || c.OidcEmailDomain != "" {
		return schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
			Type:               "openidconnect",
			ClientID:           c.OidcClientID,
			ClientSecret:       c.OidcClientSecret,
			Issuer:             c.OidcProvider,
			RequireEmailDomain: c.OidcEmailDomain,
		}}
	}
	if c.SamlIDProviderMetadataURL != "" || c.SamlSPCert != "" || c.SamlSPKey != "" {
		return schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
			Type: "saml",
			IdentityProviderMetadataURL: c.SamlIDProviderMetadataURL,
			ServiceProviderCertificate:  c.SamlSPCert,
			ServiceProviderPrivateKey:   c.SamlSPKey,
		}}
	}
	return schema.AuthProviders{}
}

// AuthPublic reports whether the site is public. Currently only the builtin auth provider allows
// sites to be public. AuthPublic only returns true if auth.public (in site config) is true *and*
// the active auth provider is builtin.
func AuthPublic() bool { return authPublic(Get()) }
func authPublic(c *schema.SiteConfiguration) bool {
	return authProvider(c).Builtin != nil && c.AuthPublic
}

// AuthAllowSignup reports whether the site allows signup. Currently only the builtin auth
// provider allows signup. AuthAllowSignup returns true if auth.allowSignup is true OR if
// auth.providers' builtin provider has allowSignup true (in site config).
func AuthAllowSignup() bool { return authAllowSignup(Get()) }
func authAllowSignup(c *schema.SiteConfiguration) bool {
	p := authProvider(c).Builtin
	return p != nil && p.AllowSignup
}

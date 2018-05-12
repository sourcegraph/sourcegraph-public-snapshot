package conf

import "github.com/sourcegraph/sourcegraph/schema"

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

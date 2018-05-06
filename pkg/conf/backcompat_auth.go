package conf

import "github.com/sourcegraph/sourcegraph/schema"

// AuthProvider returns the auth.provider value and associated provider configuration, applying the
// default "builtin" if none is specified, and respecting the legacy non-auth.provider configs
// (i.e., setting "samlSPCert" will enable the SAML auth provider even if auth.provider is not set
// to "saml").
func AuthProvider() schema.AuthProviders {
	return authProvider(Get())
}

func authProvider(c *schema.SiteConfiguration) schema.AuthProviders {
	switch {
	case c.AuthProvider == "openidconnect" || authOpenIDConnect(c) != nil:
		var p schema.OpenIDConnectAuthProvider
		if o := authOpenIDConnect(c); o != nil {
			// This nil check pattern (in the other cases too) is to avoid panics if auth.provider
			// is set but the associated config object isn't. That will produce a nicer error later
			// on.
			p = *o
		}
		p.Type = "openidconnect"
		return schema.AuthProviders{Openidconnect: &p}

	case c.AuthProvider == "saml" || authSAML(c) != nil:
		var p schema.SAMLAuthProvider
		if o := authSAML(c); o != nil {
			p = *o
		}
		p.Type = "saml"
		return schema.AuthProviders{Saml: &p}

	case c.AuthProvider == "http-header" || authHTTPHeader(c) != nil:
		var p schema.HTTPHeaderAuthProvider
		if o := authHTTPHeader(c); o != nil {
			p = *o
		}
		p.Type = "http-header"
		return schema.AuthProviders{HttpHeader: &p}

	default: // including "builtin"
		return schema.AuthProviders{
			Builtin: &schema.BuiltinAuthProvider{
				Type:        "builtin",
				AllowSignup: c.AuthAllowSignup,
			},
		}
	}
}

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

// AuthPublic reports whether the site is public. Currently only the builtin auth provider allows
// sites to be public. AuthPublic only returns true if auth.public (in site config) is true *and*
// there is a builtin auth provider.
func AuthPublic() bool { return authPublic(Get()) }
func authPublic(c *schema.SiteConfiguration) bool {
	for _, p := range c.AuthProviders {
		if p.Builtin != nil && c.AuthPublic {
			return true
		}
	}
	return false
}

// AuthAllowSignup reports whether the site allows signup. Currently only the builtin auth provider
// allows signup. AuthAllowSignup returns true if auth.providers' builtin provider has allowSignup
// true (in site config).
func AuthAllowSignup() bool { return authAllowSignup(Get()) }
func authAllowSignup(c *schema.SiteConfiguration) bool {
	for _, p := range c.AuthProviders {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return false
}

package conf

import (
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
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
	case p.Github != nil:
		return p.Github.Type
	case p.Gitlab != nil:
		return p.Gitlab.Type
	default:
		return ""
	}
}

// AuthPublic reports whether the site is public. Because many core features rely on persisted user
// settings, this leads to a degraded experience for most users. As a result, for self-hosted private
// usage it is preferable for all users to have accounts. But on sourcegraph.com, allowing users to
// opt-in to accounts remains worthwhile, despite the degraded UX.
func AuthPublic() bool { return dotcom.SourcegraphDotComMode() }

// AuthAllowSignup reports whether the site allows signup. Currently only the builtin auth provider
// allows signup. AuthAllowSignup returns true if auth.providers' builtin provider has allowSignup
// true (in site config).
func AuthAllowSignup() bool { return authAllowSignup(Get()) }
func authAllowSignup(c *Unified) bool {
	for _, p := range c.AuthProviders {
		if p.Builtin != nil && p.Builtin.AllowSignup {
			return true
		}
	}
	return false
}

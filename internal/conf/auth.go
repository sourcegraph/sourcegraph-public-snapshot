package conf

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/schema"
)

var SingleUserMode = true

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
func AuthPublic() bool { return envvar.SourcegraphDotComMode() }

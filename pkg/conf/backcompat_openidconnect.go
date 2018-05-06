package conf

import "github.com/sourcegraph/sourcegraph/schema"

// authOpenIDConnect returns the OpenIDConnectAuthProvider from the config (if enabled), regardless
// of whether the old oidc* or new auth.openidconnect properties were used.
func authOpenIDConnect(input *schema.SiteConfiguration) (p *schema.OpenIDConnectAuthProvider) {
	// oidc* properties first (lower precedence)
	if input.OidcClientID != "" || input.OidcClientSecret != "" || input.OidcProvider != "" || input.OidcEmailDomain != "" {
		p = &schema.OpenIDConnectAuthProvider{
			ClientID:           input.OidcClientID,
			ClientSecret:       input.OidcClientSecret,
			Issuer:             input.OidcProvider,
			RequireEmailDomain: input.OidcEmailDomain,
		}
	}

	// auth.openIDConnect next (higher precedence)
	if input.AuthProvider == "openidconnect" && input.AuthOpenIDConnect != nil {
		p = input.AuthOpenIDConnect
	}

	if p != nil {
		p.Type = "openidconnect"
	}

	return p
}

package conf

import "sourcegraph.com/sourcegraph/sourcegraph/schema"

// AuthOpenIDConnect returns the OpenIDConnectAuthProvider, regardless of whether
// the old oidc* or new auth.openidconnect properties were used.
func AuthOpenIDConnect() *schema.OpenIDConnectAuthProvider { return authOpenIDConnect(cfg) }

func authOpenIDConnect(input schema.SiteConfiguration) (p *schema.OpenIDConnectAuthProvider) {
	// oidc* properties first (lower precedence)
	if input.OidcClientID != "" || input.OidcClientSecret != "" || input.OidcProvider != "" || input.OidcEmailDomain != "" || input.OidcOverrideToken != "" {
		p = &schema.OpenIDConnectAuthProvider{
			ClientID:           input.OidcClientID,
			ClientSecret:       input.OidcClientSecret,
			Issuer:             input.OidcProvider,
			RequireEmailDomain: input.OidcEmailDomain,
			OverrideToken:      input.OidcOverrideToken,
		}
	}

	// auth.openIDConnect next (higher precedence)
	if input.AuthOpenIDConnect != nil {
		p = input.AuthOpenIDConnect
	}

	return p
}

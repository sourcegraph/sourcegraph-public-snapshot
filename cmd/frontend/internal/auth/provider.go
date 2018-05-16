package auth

// A Provider describes a user authentication provider (which provides functionality related to
// signing in and signing up, user identity, etc.).
//
// For example, OpenID Connect (in the ./openidconnect package) and SAML (in the ./saml package)
// implement authentication providers. These provider implementations allow users to authenticate
// using their respective authentication schemes (if enabled in site config).
//
// An authentication provider implementation can have multiple Provider instances. For example, a
// site may support OpenID Connect authentication either via G Suite or Okta, each of which would be
// represented by its own Provider instance.
type Provider struct {
	// Public is publicly visible information about this provider.
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	Public PublicProviderInfo
}

// PublicProviderInfo is publicly visible information about an auth provider.
type PublicProviderInfo struct {
	// DisplayName is the human-readable name of the authentication provider instance.
	//
	// For example, if this provider instance provides OpenID Connect via G Suite, the name should
	// mention G Suite and the associated domain name.
	DisplayName string `json:"displayName,omitempty"`

	// IsBuiltin is whether this value describes the builtin auth provider (which requires entering
	// a username and password form submission to sign in, not navigating to a login URL).
	IsBuiltin bool `json:"isBuiltin"`

	// LoginURL is the URL to visit in order to initiate authenticating via this provider.
	LoginURL string `json:"loginURL,omitempty"`
}

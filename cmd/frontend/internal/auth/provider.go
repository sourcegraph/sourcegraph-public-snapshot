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
	// ProviderID uniquely identifies this provider among all of the active providers.
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	ProviderID

	// Public is publicly visible information about this provider.
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	Public PublicProviderInfo
}

// ProviderID uniquely identifies a provider among all of the active providers.
//
// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated and
// anonymous clients.
type ProviderID struct {
	// ServiceType is the type of this auth provider's external service.
	ServiceType string `json:"serviceType"`

	// Key is a unique key among all other providers with the same ServiceType value. It must be a
	// valid single URI path component (i.e., it must not contain '/' or any other character that is
	// not valid in a URI path component).
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	Key string `json:"key,omitempty"`
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

	// AuthenticationURL is the URL to visit in order to initiate authenticating via this provider.
	//
	// TODO(sqs): Support "return-to" post-authentication-redirect destinations so newly authed
	// users aren't dumped back onto the homepage.
	AuthenticationURL string `json:"authenticationURL,omitempty"`
}

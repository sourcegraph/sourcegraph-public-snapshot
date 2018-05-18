package auth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/schema"
)

// A Provider represents a user authentication provider (which provides functionality related to
// signing in and signing up, user identity, etc.) that is present in the site configuration
// "auth.providers" array.
//
// An authentication provider implementation can have multiple Provider instances. For example, a
// site may support OpenID Connect authentication either via G Suite or Okta, each of which would be
// represented by its own Provider instance.
type Provider interface {
	// ID uniquely identifies this provider among all of the active providers.
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	ID() ProviderID

	// Config is the entry in the site configuration "auth.providers" array that this provider
	// represents.
	//
	// 🚨 SECURITY: This value contains secret information that must not be shown to
	// non-site-admins.
	Config() schema.AuthProviders

	// CachedInfo returns cached information about the provider.
	CachedInfo() *ProviderInfo

	// Refresh refreshes the provider's information with an external service, if any.
	Refresh(ctx context.Context) error
}

// ProviderID uniquely identifies a provider among all of the active providers.
//
// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated and
// anonymous clients.
type ProviderID struct {
	// Type is the type of this auth provider (equal to its "type" property in its entry in the
	// "auth.providers" array in site configuration).
	Type string

	// ID is an identifier that is unique among all other providers with the same Type value.
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	ID string
}

// ProviderInfo contains information about an authentication provider.
type ProviderInfo struct {
	// DisplayName returns the name to use when displaying the provider in the UI.
	DisplayName string

	// AuthenticationURL is the URL to visit in order to initiate authenticating via this provider.
	//
	// TODO(sqs): Support "return-to" post-authentication-redirect destinations so newly authed
	// users aren't dumped back onto the homepage.
	AuthenticationURL string
}

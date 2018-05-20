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
	// ConfigID returns the identifier for this provider's config in the auth.providers site
	// configuration array.
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	ConfigID() ProviderConfigID

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

// ProviderConfigID identifies a provider config object in the auth.providers site configuration
// array.
//
// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated and
// anonymous clients.
type ProviderConfigID struct {
	// Type is the type of this auth provider (equal to its "type" property in its entry in the
	// auth.providers array in site configuration).
	Type string

	// ID is an identifier that uniquely represents a provider's config among all other provider
	// configs of the same type.
	//
	// This value MUST NOT be persisted or used to associate accounts with this provider because it
	// can change when any property in this provider's config changes, even when those changes are
	// not material for identification (such as changing the display name).
	//
	// 🚨 SECURITY: This MUST NOT contain secret information because it is shown to unauthenticated
	// and anonymous clients.
	ID string
}

// ProviderInfo contains information about an authentication provider.
type ProviderInfo struct {
	// ServiceID identifies the external service that this authentication provider represents. It is
	// a stable identifier.
	ServiceID string

	// ClientID identifies the external service client used when communicating with the external
	// service. It is a stable identifier.
	ClientID string

	// DisplayName is the name to use when displaying the provider in the UI.
	DisplayName string

	// AuthenticationURL is the URL to visit in order to initiate authenticating via this provider.
	//
	// TODO(sqs): Support "return-to" post-authentication-redirect destinations so newly authed
	// users aren't dumped back onto the homepage.
	AuthenticationURL string
}

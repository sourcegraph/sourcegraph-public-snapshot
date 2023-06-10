// Package auth provides the Authenticator interface, which can be used to add
// authentication data to an outbound HTTP request, and concrete implementations
// for the commonly used authentication types.
package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// Authenticator instances mutate an outbound request to add whatever headers or
// other modifications are required to authenticate using the concrete type
// represented by the Authenticator. (For example, an OAuth token, or a username
// and password combination.)
//
// Note that, while Authenticate provides generic functionality, the concrete
// types should be careful to provide some method for external services to
// retrieve the values within so that unusual authentication flows can be
// supported.
type Authenticator interface {
	// Authenticate mutates the given request to include authentication
	// representing this value. In general, this will take the form of adding
	// headers.
	Authenticate(*http.Request) error

	// Hash uniquely identifies the authenticator for use in internal caching.
	// This value must use a cryptographic hash (for example, SHA-256).
	Hash() string
}

type Refreshable interface {
	// NeedsRefresh returns true if the Authenticator is no longer valid and
	// needs to be refreshed, such as checking if an OAuth token is close to
	// expiry or already expired.
	NeedsRefresh() bool

	// Refresh refreshes the Authenticator. This should be an in-place mutation,
	// and if any storage updates should happen after refreshing, that is done
	// here as well.
	Refresh(context.Context, httpcli.Doer) error
}

type AuthenticatorWithRefresh interface {
	Authenticator
	Refreshable
}

// AuthenticatorWithSSH wraps the Authenticator interface and augments it by
// additional methods to authenticate over SSH with this credential, in addition
// to the enclosed Authenticator. This can be used for a credential that needs
// to access an HTTP API, and git over SSH, for example.
type AuthenticatorWithSSH interface {
	Authenticator

	// SSHPrivateKey returns an RSA private key, and the passphrase securing it.
	SSHPrivateKey() (privateKey string, passphrase string)
	// SSHPublicKey returns the public key counterpart to the private key in OpenSSH
	// authorized_keys file format. This is usually accepted by code hosts to
	// allow access to git over SSH.
	SSHPublicKey() (publicKey string)
}

// URLAuthenticator instances allow adding credentials to URLs.
type URLAuthenticator interface {
	// SetURLUser authenticates the provided URL by modifying the User property
	// of the URL in-place.
	SetURLUser(*url.URL)
}

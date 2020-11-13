// auth provides the Authenticator interface, which can be used to add
// authentication data to an outbound HTTP request, and concrete implementations
// for the commonly used authentication types.
package auth

import "net/http"

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

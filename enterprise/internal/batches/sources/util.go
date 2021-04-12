package sources

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

// UnsupportedAuthenticatorError is returned by WithAuthenticator if the
// authenticator isn't supported on that code host.
type UnsupportedAuthenticatorError struct {
	have   string
	source string
}

func (e UnsupportedAuthenticatorError) Error() string {
	return fmt.Sprintf("authenticator type unsupported for %s sources: %s", e.source, e.have)
}

func newUnsupportedAuthenticatorError(source string, a auth.Authenticator) UnsupportedAuthenticatorError {
	return UnsupportedAuthenticatorError{
		have:   fmt.Sprintf("%T", a),
		source: source,
	}
}

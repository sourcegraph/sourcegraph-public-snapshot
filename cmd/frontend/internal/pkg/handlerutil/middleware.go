package handlerutil

import (
	"net/http"

	"github.com/gorilla/csrf"
)

// CSRFMiddleware is HTTP middleware that helps prevent cross-site request forgery. To make your
// forms compliant you will have to submit the token via the X-Csrf-Token header, which is made
// available in the client-side context.
func CSRFMiddleware(next http.Handler, secure bool) http.Handler {
	return csrf.Protect(
		[]byte("e953612ddddcdd5ec60d74e07d40218c"),
		// We do not use the name csrf_token since it is a common name. This
		// leads to conflicts between apps on localhost. See
		// https://github.com/sourcegraph/sourcegraph/issues/65
		csrf.CookieName("sg_csrf_token"),
		csrf.Path("/"),
		csrf.Secure(secure),
	)(next)
}

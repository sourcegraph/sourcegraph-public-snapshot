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
		csrf.CookieName("csrf_token"),
		csrf.Path("/"),
		csrf.Secure(secure),
	)(next)
}

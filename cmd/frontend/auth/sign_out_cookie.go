package auth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/session"
)

const SignOutCookie = session.SignOutCookie

// HasSignOutCookie returns true if the given request has a sign-out cookie.
func HasSignOutCookie(r *http.Request) bool {
	return session.HasSignOutCookie(r)
}

// SetSignOutCookie sets a sign-out cookie on the given response.
func SetSignOutCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   SignOutCookie,
		Value:  "true",
		Secure: true,
		Path:   "/",
	})
}

pbckbge buth

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/session"
)

const SignOutCookie = session.SignOutCookie

// HbsSignOutCookie returns true if the given request hbs b sign-out cookie.
func HbsSignOutCookie(r *http.Request) bool {
	return session.HbsSignOutCookie(r)
}

// SetSignOutCookie sets b sign-out cookie on the given response.
func SetSignOutCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Nbme:   SignOutCookie,
		Vblue:  "true",
		Secure: true,
		Pbth:   "/",
	})
}

package oauth2client

import (
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

const (
	// nonceCookieName is the name of the nonce cookie used as the
	// hard-to-guess random value passed to the OAuth2 state parameter,
	// which is used to prevent CSRF during OAuth2 logins.
	nonceCookieName = "session"
)

// writeNonceCookie sets a nonce cookie in the user's browser and
// returns the cookie's value (the nonce).
func writeNonceCookie(w http.ResponseWriter, r *http.Request, nonceCookiePath string) (string, error) {
	nonce := randstring.NewLen(32)
	http.SetCookie(w, &http.Cookie{
		Name:    nonceCookieName,
		Value:   nonce,
		Path:    nonceCookiePath,
		Expires: time.Now().Add(10 * time.Minute),
	})
	return nonce, nil
}

// deleteNonceCookie deletes the nonce cookie by sending a
// Set-Cookie header in the HTTP response to immediately expire it on
// the client.
func deleteNonceCookie(w http.ResponseWriter, nonceCookiePath string) {
	http.SetCookie(w, &http.Cookie{
		Name:   nonceCookieName,
		Path:   nonceCookiePath,
		MaxAge: -1,
	})
}

// nonceFromCookie returns the nonce supplied by the HTTP client in
// the request cookies, if present.
func nonceFromCookie(r *http.Request) (value string, present bool) {
	ck, err := r.Cookie(nonceCookieName)
	if err != nil {
		return "", false
	}
	return ck.Value, true
}

package oauth2client

import (
	"net/http"
	"time"
)

func addNonceCookie(r *http.Request, nonce string) {
	r.AddCookie(&http.Cookie{
		Name:    nonceCookieName,
		Value:   nonce,
		Path:    nonceCookiePath,
		Expires: time.Now().Add(10 * time.Minute),
	})
}

func nonceFromResponseCookie(r *http.Response) (value string, present bool) {
	for _, ck := range r.Cookies() {
		if ck.Name == nonceCookieName {
			return ck.Value, true
		}
	}
	return "", false
}

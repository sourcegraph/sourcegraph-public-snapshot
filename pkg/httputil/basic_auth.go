package httputil

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuth wraps h, requiring HTTP Basic auth with the given
// username and password.
func BasicAuth(username, password string, noAuthStatus int, h http.Handler) http.Handler {
	if noAuthStatus == 0 {
		noAuthStatus = http.StatusUnauthorized
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="appdash"`)
			http.Error(w, "authentication required", noAuthStatus)
			return
		}

		// Constant time comparison to avoid timing attack.
		if subtle.ConstantTimeCompare([]byte(username+password), []byte(u+p)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="appdash"`)
			http.Error(w, "bad credentials", noAuthStatus)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// CloneRequest returns a clone of the provided *http.Request. The clone is a
// shallow copy of the struct and its Header map.
func CloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}
	return r2
}

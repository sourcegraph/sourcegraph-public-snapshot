package httputil

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuthTransport is an HTTP transport that adds HTTP Basic
// authentication headers to requests.
type BasicAuthTransport struct {
	Username, Password string // username and password to authenticate with

	// Transport is the underlying HTTP transport to use when making
	// requests.  It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var transport http.RoundTripper
	if t.Transport != nil {
		transport = t.Transport
	} else {
		transport = http.DefaultTransport
	}

	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req = CloneRequest(req)
	req.SetBasicAuth(t.Username, t.Password)

	// Make the HTTP request.
	return transport.RoundTrip(req)
}

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
			http.Error(w, "bad credentials", http.StatusForbidden)
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

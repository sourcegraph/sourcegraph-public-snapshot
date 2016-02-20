package papertrail

import (
	"errors"
	"net/http"
)

// TokenTransport authenticates HTTP requests to the Papertrail API by sending
// an X-Papertrail-Token header with the provided API token (which you can
// obtain from https://papertrailapp.com/user/edit).
type TokenTransport struct {
	// Token is a valid Papertrail API token (which you can obtain from
	// https://papertrailapp.com/user/edit).
	Token string

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Token == "" {
		return nil, errors.New("Papertrail API token is empty")
	}

	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req = cloneRequest(req)
	req.Header.Set("x-papertrail-token", t.Token)

	var u http.RoundTripper
	if t.Transport != nil {
		u = t.Transport
	} else {
		u = http.DefaultTransport
	}

	return u.RoundTrip(req)
}

// Client returns an *http.Client that makes requests which are authenticated
// using the TokenTransport.
func (t *TokenTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// cloneRequest returns a clone of the provided *http.Request. The clone is a
// shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
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

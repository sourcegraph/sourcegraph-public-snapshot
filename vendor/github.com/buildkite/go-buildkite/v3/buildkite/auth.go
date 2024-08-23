package buildkite

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// TokenAuthTransport manages injection of the API token for each request
type TokenAuthTransport struct {
	APIToken  string
	APIHost   string
	Transport http.RoundTripper
}

// RoundTrip invoked each time a request is made
func (t TokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == t.APIHost || t.APIHost == "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.APIToken))
	}
	return t.transport().RoundTrip(req)
}

// Client builds a new http client.
func (t *TokenAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *TokenAuthTransport) transport() http.RoundTripper {
	// Use the custom transport if one was provided
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}

// NewTokenConfig configure authentication using an API token
// NOTE: the debug flag is not used anymore.
func NewTokenConfig(apiToken string, debug bool) (*TokenAuthTransport, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("Invalid token, empty string supplied")
	}
	return &TokenAuthTransport{APIToken: apiToken}, nil
}

// BasicAuthTransport manages injection of the authorization header
type BasicAuthTransport struct {
	APIHost  string
	Username string
	Password string
}

// RoundTrip invoked each time a request is made
func (bat BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == bat.APIHost || bat.APIHost == "" {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
				bat.Username, bat.Password)))))
	}
	return http.DefaultTransport.RoundTrip(req)
}

// Client builds a new http client.
func (bat *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: bat}
}

// NewBasicConfig configure authentication using the supplied credentials
func NewBasicConfig(username string, password string) (*BasicAuthTransport, error) {
	if username == "" {
		return nil, fmt.Errorf("Invalid username, empty string supplied")
	}
	if password == "" {
		return nil, fmt.Errorf("Invalid password, empty string supplied")
	}
	return &BasicAuthTransport{"", username, password}, nil
}

package core

import (
	"crypto/tls"
	"net/http"
	"time"
)

// An Client performs http.Requests. It captures the Do
// method of the http.Client.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientOptions are options that can be set on the HTTP client.
type ClientOptions struct {
	APIToken string

	Cert       string
	CertUser   string
	SessionKey string

	InsecureSkipVerify bool
	Timeout            time.Duration

	// If set, Client will be used to execute HTTP requests.
	// Otherwise, one is created with default settings and
	// InsecureSkipVerify respected.
	Client Client
}

// makeHttpClient creates a new HTTP client for making API requests.
func makeHTTPClient(options *ClientOptions) Client {
	if options.Client != nil {
		return options.Client
	}

	return &http.Client{
		Timeout: options.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: options.InsecureSkipVerify,
			},
		},
	}
}

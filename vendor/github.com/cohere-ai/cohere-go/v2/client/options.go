// This file was auto-generated by Fern from our API Definition.

package client

import (
	core "github.com/cohere-ai/cohere-go/v2/core"
	option "github.com/cohere-ai/cohere-go/v2/option"
	http "net/http"
)

// WithBaseURL sets the base URL, overriding the default
// environment, if any.
func WithBaseURL(baseURL string) *core.BaseURLOption {
	return option.WithBaseURL(baseURL)
}

// WithHTTPClient uses the given HTTPClient to issue the request.
func WithHTTPClient(httpClient core.HTTPClient) *core.HTTPClientOption {
	return option.WithHTTPClient(httpClient)
}

// WithHTTPHeader adds the given http.Header to the request.
func WithHTTPHeader(httpHeader http.Header) *core.HTTPHeaderOption {
	return option.WithHTTPHeader(httpHeader)
}

// WithMaxAttempts configures the maximum number of retry attempts.
func WithMaxAttempts(attempts uint) *core.MaxAttemptsOption {
	return option.WithMaxAttempts(attempts)
}

// WithToken sets the 'Authorization: Bearer <token>' request header.
func WithToken(token string) *core.TokenOption {
	return option.WithToken(token)
}

// WithClientName sets the clientName request header.
//
// The name of the project that is making the request.
func WithClientName(clientName *string) *core.ClientNameOption {
	return option.WithClientName(clientName)
}

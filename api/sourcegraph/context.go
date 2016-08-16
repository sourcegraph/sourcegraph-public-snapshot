package sourcegraph

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"context"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type contextKey int

const (
	grpcEndpointKey contextKey = iota
	mockClientKey
	credentialsKey
	clientMetadataKey
)

// WithGRPCEndpoint returns a copy of parent whose clients (obtained
// using FromContext) communicate with the given gRPC API endpoint
// URL.
func WithGRPCEndpoint(parent context.Context, url *url.URL) context.Context {
	return context.WithValue(parent, grpcEndpointKey, url)
}

// GRPCEndpoint returns the context's gRPC endpoint URL that was
// previously configured using WithGRPCEndpoint.
func GRPCEndpoint(ctx context.Context) *url.URL {
	url, _ := ctx.Value(grpcEndpointKey).(*url.URL)
	if url == nil {
		panic("no gRPC API endpoint URL set in context")
	}
	return url
}

// WithMockClient returns a copy of parent whose clients (obtained using
// FromContext) communicate with the given mock client. NewClientFromContext
// checks this value and returns it if it is set.
func WithMockClient(parent context.Context, client *Client) context.Context {
	return context.WithValue(parent, mockClientKey, client)
}

// MockClient returns the context's mocked client if it was
// previously set using WithMockClient.
func MockClient(ctx context.Context) *Client {
	client, ok := ctx.Value(mockClientKey).(*Client)
	if !ok {
		return nil
	}
	return client
}

// Credentials authenticate gRPC requests made by an API client.
type Credentials interface {
	oauth2.TokenSource
}

// WithCredentials returns a copy of the parent context that uses cred
// as the credentials for future API clients constructed using this
// context (with NewClientFromContext). It replaces (shadows) any
// previously set credentials in the context.
//
// It can be used to add, e.g., trace/span ID metadata for request
// tracing.
func WithCredentials(parent context.Context, cred Credentials) context.Context {
	return context.WithValue(parent, credentialsKey, cred)
}

// CredentialsFromContext returns the credentials (if any) previously
// set in the context by WithCredentials.
func CredentialsFromContext(ctx context.Context) Credentials {
	cred, ok := ctx.Value(credentialsKey).(Credentials)
	if !ok {
		return nil
	}
	return cred
}

// WithClientMetadata returns a copy of the parent context that merges
// in the specified metadata to future API clients constructed using
// this context (with NewClientFromContext). It replaces (shadows) any
// previously set metadata in the context.
func WithClientMetadata(parent context.Context, md map[string]string) context.Context {
	return context.WithValue(parent, clientMetadataKey, md)
}

// clientMetadataFromContext returns the metadata (if any) previously
// set in the context by WithClientMetadata.
func clientMetadataFromContext(ctx context.Context) map[string]string {
	cred, ok := ctx.Value(clientMetadataKey).(map[string]string)
	if !ok {
		return nil
	}
	return cred
}

// NewClientFromContext returns a Sourcegraph API client that
// communicates with the Sourcegraph gRPC endpoint in ctx (i.e.,
// GRPCEndpoint(ctx)).
func NewClientFromContext(ctx context.Context) (*Client, error) {
	newClientFromContextMu.RLock()
	f := newClientFromContext
	newClientFromContextMu.RUnlock()
	return f(ctx)
}

// MockNewClientFromContext allows a test to mock out the return value of
// NewClientFromContext. Note that this is modifying global state, so if your
// tests run in parallel you may get unexpected results
func MockNewClientFromContext(f func(ctx context.Context) (*Client, error)) {
	newClientFromContextMu.Lock()
	newClientFromContext = f
	newClientFromContextMu.Unlock()
}

// RestoreNewClientFromContext removes the mock and returns the correct
// implementation
func RestoreNewClientFromContext() {
	newClientFromContextMu.Lock()
	newClientFromContext = realNewClientFromContext
	newClientFromContextMu.Unlock()
}

var (
	maxDialTimeout         = 10 * time.Second
	newClientFromContextMu sync.RWMutex
	newClientFromContext   = realNewClientFromContext
)
var realNewClientFromContext = func(ctx context.Context) (*Client, error) {
	mockClient := MockClient(ctx)
	if mockClient != nil {
		return mockClient, nil
	}

	opts := []grpc.DialOption{
		grpc.WithCodec(GRPCCodec),
	}

	grpcEndpoint := GRPCEndpoint(ctx)
	if grpcEndpoint.Scheme == "https" {
		creds := credentials.NewClientTLSFromCert(nil, "")
		if host, _, _ := net.SplitHostPort(grpcEndpoint.Host); host == "localhost" {
			creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Use contextCredentials instead of directly using the cred
	// so that we can use different credentials for the same
	// connection (in the pool).
	opts = append(opts, grpc.WithPerRPCCredentials(contextCredentials{}))

	// Dial timeout is the lesser of the ctx deadline or
	// maxDialTimeout.
	var timeout time.Duration
	if d, ok := ctx.Deadline(); ok && time.Now().Add(maxDialTimeout).After(d) {
		timeout = d.Sub(time.Now())
	} else {
		timeout = maxDialTimeout
	}
	opts = append(opts, grpc.WithTimeout(timeout))

	conn, err := pooledGRPCDial(hostWithExplicitPort(grpcEndpoint), opts...)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

// hostWithExplicitPort returns u's host with an explicit port number
// (determined by the scheme), if none is present.
func hostWithExplicitPort(u *url.URL) string {
	if _, _, err := net.SplitHostPort(u.Host); err != nil && strings.Contains(err.Error(), "missing port in address") {
		var port int
		switch u.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		}
		return fmt.Sprintf("%s:%d", u.Host, port)
	}
	return u.Host
}

// RemovePooledGRPCConn removes the pooled grpc.ClientConnection to the gRPC endpoint
// in the context. The result of calling this function  is that the pooled connection for
// this endpoint will be reset, so the subsequent call to NewClientFromContext() would have
// to dial a new gRPC connection to this endpoint.
var RemovePooledGRPCConn = func(ctx context.Context) {
	grpcEndpoint := GRPCEndpoint(ctx)
	removeConnFromPool(grpcEndpoint.Host)
}

// contextCredentials implements the credentials.Credentials interface.
type contextCredentials struct{}

// GetRequestMetadata implements the credentials.Credentials interface. As per the
// interface definition, it may be called by multiple goroutines concurrently.
func (contextCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	m := clientMetadataFromContext(ctx)

	if cred := CredentialsFromContext(ctx); cred != nil {
		credMD, err := (oauth.TokenSource{TokenSource: cred}).GetRequestMetadata(ctx)
		if err != nil {
			return nil, err
		}

		if m == nil {
			m = credMD
		} else {
			// Copy the map to avoid a data race writing to the map returned by
			// clientMetaDataFromContext.
			cpy := make(map[string]string)
			for k, v := range m {
				cpy[k] = v
			}
			for k, v := range credMD {
				cpy[k] = v
			}
			m = cpy
		}
	}
	return m, nil
}

// RequireTransportSecurity implements the credentials.Credentials interface.
func (contextCredentials) RequireTransportSecurity() bool {
	return false
}

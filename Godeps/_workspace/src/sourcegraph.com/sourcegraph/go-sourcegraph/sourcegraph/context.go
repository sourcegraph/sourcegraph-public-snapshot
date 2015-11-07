package sourcegraph

import (
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type contextKey int

const (
	grpcEndpointKey contextKey = iota
	httpEndpointKey
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

var maxDialTimeout = 10 * time.Second

// NewClientFromContext returns a Sourcegraph API client that
// communicates with the Sourcegraph gRPC endpoint in ctx (i.e.,
// GRPCEndpoint(ctx)).
var NewClientFromContext = func(ctx context.Context) *Client {
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

	conn, err := pooledGRPCDial(grpcEndpoint.Host, opts...)
	if err != nil {
		panic(err)
	}
	c := NewClient(conn)
	return c
}

// RemovePooledGRPCConn removes the pooled grpc.ClientConnection to the gRPC endpoint
// in the context. The result of calling this function  is that the pooled connection for
// this endpoint will be reset, so the subsequent call to NewClientFromContext() would have
// to dial a new gRPC connection to this endpoint.
var RemovePooledGRPCConn = func(ctx context.Context) {
	grpcEndpoint := GRPCEndpoint(ctx)
	removeConnFromPool(grpcEndpoint.Host)
}

type contextCredentials struct{}

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

func (contextCredentials) RequireTransportSecurity() bool {
	return false
}

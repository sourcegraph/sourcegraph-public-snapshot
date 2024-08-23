// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package auth // import "go.opentelemetry.io/collector/extension/auth"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// Server is an Extension that can be used as an authenticator for the configauth.Authentication option.
// Authenticators are then included as part of OpenTelemetry Collector builds and can be referenced by their
// names from the Authentication configuration. Each Server is free to define its own behavior and configuration options,
// but note that the expectations that come as part of Extensions exist here as well. For instance, multiple instances of the same
// authenticator should be possible to exist under different names.
type Server interface {
	extension.Extension

	// Authenticate checks whether the given headers map contains valid auth data. Successfully authenticated calls will always return a nil error.
	// When the authentication fails, an error must be returned and the caller must not retry. This function is typically called from interceptors,
	// on behalf of receivers, but receivers can still call this directly if the usage of interceptors isn't suitable.
	// The deadline and cancellation given to this function must be respected, but note that authentication data has to be part of the map, not context.
	// The resulting context should contain the authentication data, such as the principal/username, group membership (if available), and the raw
	// authentication data (if possible). This will allow other components in the pipeline to make decisions based on that data, such as routing based
	// on tenancy as determined by the group membership, or passing through the authentication data to the next collector/backend.
	// The context keys to be used are not defined yet.
	Authenticate(ctx context.Context, headers map[string][]string) (context.Context, error)
}

type defaultServer struct {
	ServerAuthenticateFunc
	component.StartFunc
	component.ShutdownFunc
}

// ServerOption represents the possible options for NewServer.
type ServerOption func(*defaultServer)

// ServerAuthenticateFunc defines the signature for the function responsible for performing the authentication based
// on the given headers map. See Server.Authenticate.
type ServerAuthenticateFunc func(ctx context.Context, headers map[string][]string) (context.Context, error)

func (f ServerAuthenticateFunc) Authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	if f == nil {
		return ctx, nil
	}
	return f(ctx, headers)
}

// WithServerAuthenticate specifies which function to use to perform the authentication.
func WithServerAuthenticate(authFunc ServerAuthenticateFunc) ServerOption {
	return func(o *defaultServer) {
		o.ServerAuthenticateFunc = authFunc
	}
}

// WithServerStart overrides the default `Start` function for a component.Component.
// The default always returns nil.
func WithServerStart(startFunc component.StartFunc) ServerOption {
	return func(o *defaultServer) {
		o.StartFunc = startFunc
	}
}

// WithServerShutdown overrides the default `Shutdown` function for a component.Component.
// The default always returns nil.
func WithServerShutdown(shutdownFunc component.ShutdownFunc) ServerOption {
	return func(o *defaultServer) {
		o.ShutdownFunc = shutdownFunc
	}
}

// NewServer returns a Server configured with the provided options.
func NewServer(options ...ServerOption) Server {
	bc := &defaultServer{}

	for _, op := range options {
		op(bc)
	}

	return bc
}

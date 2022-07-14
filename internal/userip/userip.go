package userip

import (
	"context"
)

type userIPKey struct{}

type UserIP struct {
	// IP identifies the IP of the client.
	IP string
	// XForwardedFor identifies the originating IP address of a client.
	XForwardedFor string
}

// FromContext retrieves UserIP, if available, from context.
func FromContext(ctx context.Context) *UserIP {
	ip, ok := ctx.Value(userIPKey{}).(*UserIP)
	if !ok || ip == nil {
		return nil
	}
	return ip
}

// WithUserIP adds user IP information to context for propagation.
func WithUserIP(ctx context.Context, userIP *UserIP) context.Context {
	return context.WithValue(ctx, userIPKey{}, userIP)
}

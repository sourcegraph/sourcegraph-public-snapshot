package eventsutil

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

type contextKey int

// Session is the information stored in a session cookie.
type Session struct {
	UID int
	// DeviceID is a unique identifier assigner per user (per device).
	DeviceID string
}

// sessionCookieName is the name of the session cookie.
const sessionCookieName = "session-device-id"

const (
	userAgentKey contextKey = iota
	deviceIdKey
)

// AgentMiddleware fetches the user's user agent and stores it
// in the context for downstream HTTP handlers.
func AgentMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)
	ctx = WithUserAgent(ctx, url.QueryEscape(r.UserAgent()))
	httpctx.SetForRequest(r, ctx)
	next(w, r)
}

// WithUserAgent returns a copy of the context with the user agent added to it
// (and available via UserAgentFromContext). Generally you should use
// AgentMiddleware to set it in the context; WithUserAgent is probably most
// useful for tests where you want to inject a specific user agent.
func WithUserAgent(ctx context.Context, useragent string) context.Context {
	return context.WithValue(ctx, userAgentKey, useragent)
}

// UserAgentFromContext returns the user agent from context.
func UserAgentFromContext(ctx context.Context) string {
	user, _ := ctx.Value(userAgentKey).(string)
	return user
}

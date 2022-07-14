package userip

import (
	"context"
	"net/http"
	"strings"
)

type userIPKey struct{}

type UserIP struct {
	IP            string
	XForwardedFor string
}

func FromContext(ctx context.Context) *UserIP {
	a, ok := ctx.Value(userIPKey{}).(*UserIP)
	if !ok || a == nil {
		return nil
	}
	return a
}

func WithUserIP(ctx context.Context, userIP *UserIP) context.Context {
	return context.WithValue(ctx, userIPKey{}, userIP)
}

const (
	// headerKeyUserIP
	headerKeyUserIP = "X-Sourcegraph-User-IP"
	// headerKeyForwardedFor
	headerKeyForwardedFor = "X-Forwarded-For"
)

type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	userIP := FromContext(req.Context())
	if userIP != nil {
		req.Header.Set(headerKeyUserIP, userIP.IP)
		req.Header.Set(headerKeyForwardedFor, userIP.XForwardedFor)
	}

	return t.RoundTripper.RoundTrip(req)
}

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var userIP UserIP
		userIP.IP = strings.Split(req.RemoteAddr, ":")[0]
		userIP.XForwardedFor = req.Header.Get(headerKeyForwardedFor)

		ctx := req.Context()
		ctxWithIP := WithUserIP(ctx, &userIP)

		next.ServeHTTP(rw, req.WithContext(ctxWithIP))
	})
}

package userip

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
)

type contextKey int

const userIPKey contextKey = iota

type UserIP string

func FromContext(ctx context.Context) UserIP {
	a, ok := ctx.Value(userIPKey).(UserIP)
	if !ok || a == "" {
		return ""
	}
	return a
}

func WithUserIP(ctx context.Context, ip UserIP) context.Context {
	return context.WithValue(ctx, userIPKey, ip)
}

const (
	// headerKeyUserIP
	headerKeyUserIP = "X-Sourcegraph-User-IP"
)

type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	ip := FromContext(req.Context())
	req.Header.Set(headerKeyUserIP, string(ip))

	return t.RoundTripper.RoundTrip(req)
}

func UserIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var ip UserIP
		if fIp := req.Header.Get("X-FORWARDED-FOR"); fIp != "" {
			ip = UserIP(fIp)
		} else {
			ip = UserIP(req.RemoteAddr)
		}
		ctx := req.Context()
		ctxWithIP := WithUserIP(ctx, ip)

		next.ServeHTTP(rw, req.WithContext(ctxWithIP))
	})
}

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userIP := req.Header.Get(headerKeyUserIP)
		log.Scoped("userip", "logging user ip").Info("userip", log.String("path", req.URL.Path), log.String("ip", userIP))
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

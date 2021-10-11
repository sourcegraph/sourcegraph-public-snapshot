package deviceid

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/cookie"
)

type deviceIdKey struct{}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(contextWithDeviceID(r)))
	})
}

func contextWithDeviceID(r *http.Request) context.Context {
	if deviceId, ok := cookie.DeviceID(r); ok {
		return context.WithValue(r.Context(), deviceIdKey{}, deviceId)
	}
	return r.Context()
}

func FromContext(ctx context.Context) string {
	if deviceId := ctx.Value(deviceIdKey{}); deviceId != nil {
		return deviceId.(string)
	}
	return ""
}

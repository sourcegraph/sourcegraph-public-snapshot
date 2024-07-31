package deviceid

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
)

type deviceIDKey struct{}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		if _, ok := cookie.DeviceID(r); !ok {
			newDeviceId, _ := uuid.NewRandom()
			http.SetCookie(w, &http.Cookie{
				Name:    "sourcegraphDeviceId",
				Value:   newDeviceId.String(),
				Expires: time.Now().AddDate(1, 0, 0),
				Secure:  conf.ExternalURLParsed().Scheme == "https",
				Domain:  r.URL.Host,
			})
		}
		next.ServeHTTP(w, r.WithContext(contextWithDeviceID(r)))
	})
}

func contextWithDeviceID(r *http.Request) context.Context {
	if deviceID, ok := cookie.DeviceID(r); ok {
		return context.WithValue(r.Context(), deviceIDKey{}, deviceID)
	}

	return r.Context()
}

func FromContext(ctx context.Context) string {
	if deviceID := ctx.Value(deviceIDKey{}); deviceID != nil {
		return deviceID.(string)
	}
	return ""
}

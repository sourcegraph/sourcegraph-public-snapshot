package confauth

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

func setAllowAnonymousUsageContextKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if info, err := licensing.GetConfiguredProductLicenseInfo(); err == nil && info != nil {
			ctx = context.WithValue(r.Context(), auth.AllowAnonymousRequestContextKey, info.HasTag(licensing.AllowAnonymousUsageTag))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Middleware() *auth.Middleware {
	return &auth.Middleware{
		API: setAllowAnonymousUsageContextKey,
		App: setAllowAnonymousUsageContextKey,
	}
}

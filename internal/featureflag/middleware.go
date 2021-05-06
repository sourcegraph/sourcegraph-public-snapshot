package featureflag

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type featureFlagKey struct{}

func FeatureFlagMiddleware(ffs *database.FeatureFlagStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(contextWithFeatureFlags(ffs, r)))
	})
}

func contextWithFeatureFlags(ffs *database.FeatureFlagStore, r *http.Request) context.Context {
	if a := actor.FromContext(r.Context()); a != nil && a.UID != 0 {
		flags, err := ffs.UserFlags(r.Context(), a.UID)
		if err == nil {

			return context.WithValue(r.Context(), featureFlagKey{}, flags)
		}
		// Continue if err != nil
	}

	if cookie, err := r.Cookie("sourcegraphAnonymousUid"); err != nil {
		flags, err := ffs.AnonymousUserFlags(r.Context(), cookie.Value)
		if err == nil {
			return context.WithValue(r.Context(), featureFlagKey{}, flags)
		}
		// Continue if err != nil
	}

	flags, err := ffs.UserlessFeatureFlags(r.Context())
	if err != nil {
		return context.WithValue(r.Context(), featureFlagKey{}, flags)
	}
	return r.Context()
}

type flagContextKey struct{}

// FromContext retrieves the current set of flags from the current
// request's context.
func FromContext(ctx context.Context) FlagSet {
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		return flags.(FlagSet)
	}
	return nil
}

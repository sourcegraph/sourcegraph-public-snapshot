package featureflag

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type flagContextKey struct{}

// FeatureFlagMiddleware evaluates the feature flags for the current user and adds the
// feature flags to the current context.
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
			return context.WithValue(r.Context(), flagContextKey{}, FlagSet(flags))
		}
		// Continue if err != nil
	}

	if cookie, err := r.Cookie("sourcegraphAnonymousUid"); err != nil {
		flags, err := ffs.AnonymousUserFlags(r.Context(), cookie.Value)
		if err == nil {
			return context.WithValue(r.Context(), flagContextKey{}, FlagSet(flags))
		}
		// Continue if err != nil
	}

	flags, err := ffs.UserlessFeatureFlags(r.Context())
	if err != nil {
		return r.Context()
	}
	return context.WithValue(r.Context(), flagContextKey{}, FlagSet(flags))
}

// FromContext retrieves the current set of flags from the current
// request's context.
func FromContext(ctx context.Context) FlagSet {
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		return flags.(FlagSet)
	}
	return nil
}

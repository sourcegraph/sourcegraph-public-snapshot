package featureflag

import (
	"context"
	"net/http"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/cookie"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

type flagContextKey struct{}

type Store interface {
	GetUserFlags(context.Context, int32) (map[string]bool, error)
	GetAnonymousUserFlags(context.Context, string) (map[string]bool, error)
	GetGlobalFeatureFlags(context.Context) (map[string]bool, error)
}

// Middleware evaluates the feature flags for the current user and adds the
// feature flags to the current context.
func Middleware(ffs Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(contextWithFeatureFlags(ffs, r)))
	})
}

func contextWithFeatureFlags(ffs Store, r *http.Request) context.Context {
	fetcher := &flagSetFetcher{ffs: ffs, r: r}
	return context.WithValue(r.Context(), flagContextKey{}, fetcher)
}

// flagSetFetcher is a lazy fetcher for a FlagSet, given an *http.Request. It
// will fetch the flags as required, once they're loaded from the context. This
// pattern prevents us from loading feature flags on every request, even when
// we don't end up using them.
type flagSetFetcher struct {
	r   *http.Request
	ffs Store

	once    sync.Once
	flagSet FlagSet
}

func (f *flagSetFetcher) Fetch(ctx context.Context) FlagSet {
	f.once.Do(func() {
		if a := actor.FromContext(ctx); a.IsAuthenticated() {
			flags, err := f.ffs.GetUserFlags(ctx, a.UID)
			if err == nil {
				f.flagSet = FlagSet(flags)
				return
			}
			// Continue if err != nil
		}

		if uid, ok := cookie.AnonymousUID(f.r); ok {
			flags, err := f.ffs.GetAnonymousUserFlags(ctx, uid)
			if err == nil {
				f.flagSet = FlagSet(flags)
				return
			}
			// Continue if err != nil
		}

		flags, err := f.ffs.GetGlobalFeatureFlags(ctx)
		if err == nil {
			f.flagSet = FlagSet(flags)
		}
	})

	return f.flagSet
}

// FromContext retrieves the current set of flags from the current
// request's context.
func FromContext(ctx context.Context) FlagSet {
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		return flags.(*flagSetFetcher).Fetch(ctx)
	}
	return nil
}

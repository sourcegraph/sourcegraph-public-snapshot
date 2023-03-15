package featureflag

import (
	"context"
	"net/http"
	"sync"

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

		store := ffs
		if flags, ok := requestOverrides(r); ok {
			store = &overrideStore{
				store: ffs,
				flags: flags,
			}
		}

		next.ServeHTTP(w, r.WithContext(WithFlags(r.Context(), store)))
	})
}

// flagSetFetcher is a lazy fetcher for a FlagSet. It will fetch the flags as
// required, once they're loaded from the context. This pattern prevents us
// from loading feature flags on every request, even when we don't end up using
// them.
type flagSetFetcher struct {
	ffs Store

	once sync.Once
	// Actor is the actor that was used to populate flagSet
	actor *actor.Actor
	// flagSet is the once-populated set of flags for the actor at the time of population
	flagSet *FlagSet
}

func (f *flagSetFetcher) fetch(ctx context.Context) *FlagSet {
	f.once.Do(func() {
		f.actor = actor.FromContext(ctx)
		f.flagSet = f.fetchForActor(ctx, f.actor)
	})

	currentActor := actor.FromContext(ctx)
	if f.actor == currentActor {
		// If the actor hasn't changed, return the cached flag set
		return f.flagSet
	}

	// Otherwise, re-fetch the flag set
	return f.fetchForActor(ctx, currentActor)
}

func (f *flagSetFetcher) fetchForActor(ctx context.Context, a *actor.Actor) *FlagSet {
	if a.IsAuthenticated() {
		flags, err := f.ffs.GetUserFlags(ctx, a.UID)
		if err == nil {
			return &FlagSet{flags: flags, actor: f.actor}
		}
		// Continue if err != nil
	}

	if a.AnonymousUID != "" {
		flags, err := f.ffs.GetAnonymousUserFlags(ctx, a.AnonymousUID)
		if err == nil {
			return &FlagSet{flags: flags, actor: f.actor}
		}
		// Continue if err != nil
	}

	flags, err := f.ffs.GetGlobalFeatureFlags(ctx)
	if err == nil {
		return &FlagSet{flags: flags, actor: f.actor}
	}

	return &FlagSet{actor: f.actor}
}

// FromContext retrieves the current set of flags from the current
// request's context.
func FromContext(ctx context.Context) *FlagSet {
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		return flags.(*flagSetFetcher).fetch(ctx)
	}
	return nil
}

func GetEvaluatedFlagSet(ctx context.Context) EvaluatedFlagSet {
	if flagSet := FromContext(ctx); flagSet != nil {
		return getEvaluatedFlagSetFromCache(flagSet)
	}
	return nil
}

// WithFlags adds a flag fetcher to the context so consumers of the
// returned context can use FromContext.
func WithFlags(ctx context.Context, ffs Store) context.Context {
	fetcher := &flagSetFetcher{ffs: ffs}
	return context.WithValue(ctx, flagContextKey{}, fetcher)
}

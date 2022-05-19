package featureflag

import (
	"context"
	"net/http"

	"sigs.k8s.io/kustomize/kyaml/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

type flagContextKey struct{}

//go:generate ../../dev/mockgen.sh  github.com/sourcegraph/sourcegraph/internal/featureflag -i Store -o store_mock_test.go
type Store interface {
	GetFeatureFlags(context.Context) ([]*FeatureFlag, error)
	GetUserFlag(ctx context.Context, userID int32, flagName string) (*bool, error)
	GetAnonymousUserFlag(ctx context.Context, anonymousUID string, flagName string) (*bool, error)
	GetGlobalFeatureFlag(ctx context.Context, flagName string) (*bool, error)
}

// Middleware evaluates the feature flags for the current user and adds the
// feature flags to the current context.
func Middleware(ffs Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(WithFlags(r.Context(), ffs)))
	})
}

// flagSetFetcher is a lazy fetcher for a FlagSet. It will fetch the flags as
// required, once they're loaded from the context. This pattern prevents us
// from loading feature flags on every request, even when we don't end up using
// them.
type flagSetFetcher struct {
	ffs Store
}

// TODO: add description
func (f *flagSetFetcher) evaluateForActor(ctx context.Context, a *actor.Actor, flagName string) (*bool, error) {
	if a.IsAuthenticated() {
		flag, err := f.ffs.GetUserFlag(ctx, a.UID, flagName)
		if err == nil {
			setEvaluatedFlagToCache(flagName, a, *flag)
			return flag, nil
		}
		// Continue if err != nil
	}

	if a.AnonymousUID != "" {
		flag, err := f.ffs.GetAnonymousUserFlag(ctx, a.AnonymousUID, flagName)
		if err == nil {
			setEvaluatedFlagToCache(flagName, a, *flag)
			return flag, nil
		}
		// Continue if err != nil
	}

	flag, err := f.ffs.GetGlobalFeatureFlag(ctx, flagName)
	if err == nil {
		return flag, nil
	}

	return nil, errors.Errorf("Couldn't evaluate feature flag \"%s\" for the given actor", flagName)
}

func EvaluateForActorFromContext(ctx context.Context, flagName string) (result bool) {
	result = false
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		actor := actor.FromContext(ctx)
		value, err := flags.(*flagSetFetcher).evaluateForActor(ctx, actor, flagName)
		if err == nil {
			result = *value
		}
	}
	return result
}

func GetEvaluatedFlagsFromContext(ctx context.Context) FlagSet {
	if flags := ctx.Value(flagContextKey{}); flags != nil {
		if f, err := flags.(*flagSetFetcher).ffs.GetFeatureFlags(ctx); err == nil {
			return getEvaluatedFlagSetFromCache(f, actor.FromContext(ctx))
		}
	}

	return FlagSet{}
}

// WithFlags adds a flag fetcher to the context so consumers of the
// returned context can use FromContext.
func WithFlags(ctx context.Context, ffs Store) context.Context {
	fetcher := &flagSetFetcher{ffs: ffs}
	return context.WithValue(ctx, flagContextKey{}, fetcher)
}

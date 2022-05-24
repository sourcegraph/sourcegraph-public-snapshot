package featureflag

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

// TODO
// - test for setEvaluatedFlagToCache executation
// - test for org feature flag override
// - test for user feature flag override
// - test for user & org feature flag override

func TestMiddleware(t *testing.T) {
	// Create a request with an actor on its context
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)
	req = req.WithContext(actor.WithActor(context.Background(), actor.FromUser(1)))

	mockStore := NewMockStore()
	mockStore.GetUserFlagFunc.SetDefaultHook(func(ctx context.Context, uid int32, flag string) (*bool, error) {
		if uid == 1 && flag == "test-flag" {
			value := true

			return &value, nil
		}

		return nil, nil
	})

	handler := http.Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// After going through the middleware, a request with an actor should
		// also have feature flags available.
		require.True(t, EvaluateForActorFromContext(r.Context(), "test-flag"))
	}))
	handler = Middleware(mockStore, handler)

	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGetEvaluatedFlagsFromContext(t *testing.T) {
	// TODO: test GetEvaluatedFlagsFromContext that whatever is returned by getEvaluatedFlagSetFromCache is used
}

func TestEvaluateForActorFromContext(t *testing.T) {
	// TODO: case 1: test that for authenticated user GetUserFlagFunc is called
	// TODO: case 2: test that for anonymous user GetAnonymousUserFlagFunc is called
	// TODO: case 3: test that for rest GetGlobalFeatureFlagFunc is called
	// TODO: case 4: test that for each above cases setEvaluatedFlagToCache has been called with proper args
	mockStore := NewMockStore()
	mockStore.GetUserFlagFunc.SetDefaultHook(func(_ context.Context, uid int32, flag string) (*bool, error) {
		if flag == "test-flag-1" {
			value := false

			if uid == 1 {
				value = true
			}

			return &value, nil
		}

		if flag == "test-flag-2" {
			value := false

			if uid == 2 {
				value = true
			}

			return &value, nil
		}

		return nil, nil
	})

	actor1 := actor.FromUser(1)
	actor2 := actor.FromUser(2)

	ctx := context.Background()
	ctx = actor.WithActor(ctx, actor1)
	ctx = WithFlags(ctx, mockStore)

	// Make sure user1 flags are set
	require.True(t, EvaluateForActorFromContext(ctx, "test-flag-1"))
	require.False(t, EvaluateForActorFromContext(ctx, "test-flag-2"))

	// With a new actor, the flag fetcher should re-fetch
	ctx = actor.WithActor(ctx, actor2)
	require.True(t, EvaluateForActorFromContext(ctx, "test-flag-2"))
	require.False(t, EvaluateForActorFromContext(ctx, "test-flag-1"))

	mockrequire.CalledN(t, mockStore.GetUserFlagFunc, 4)
}

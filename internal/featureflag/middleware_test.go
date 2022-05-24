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
	t.Run("authenticated user", func(t *testing.T) {
		mockStore := NewMockStore()

		mockStore.GetUserFlagFunc.SetDefaultHook(func(_ context.Context, uid int32, flag string) (*bool, error) {
			/**
			const mockFlags = {
				"1": [["f1", true]]
			}
			*/
			if flag == "f1" {
				value := false

				if uid == 1 {
					value = true
				}

				return &value, nil
			}

			if flag == "f2" {
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
		require.True(t, EvaluateForActorFromContext(ctx, "f1"))
		require.False(t, EvaluateForActorFromContext(ctx, "f2"))

		// With a new actor, the flag fetcher should re-fetch
		ctx = actor.WithActor(ctx, actor2)
		require.True(t, EvaluateForActorFromContext(ctx, "f2"))
		require.False(t, EvaluateForActorFromContext(ctx, "f1"))

		mockrequire.CalledN(t, mockStore.GetUserFlagFunc, 4)
		// TODO: test that for each above cases setEvaluatedFlagToCache has been called with proper args
	})

	// TODO: case 2: test that for anonymous user GetAnonymousUserFlagFunc is called
	t.Run("anonymous user", func(t *testing.T) {
		// actor.FromAnonymousUser("test-user")
		mockStore := NewMockStore()

		mockStore.GetAnonymousUserFlagFunc.SetDefaultHook(func(_ context.Context, anonymousUID string, flag string) (*bool, error) {
			if flag == "f1" {
				value := false

				if anonymousUID == "t1" {
					value = true
				}

				return &value, nil
			}

			if flag == "f2" {
				value := false

				if anonymousUID == "t2" {
					value = true
				}

				return &value, nil
			}

			return nil, nil
		})

		actor1 := actor.FromAnonymousUser("t1")
		actor2 := actor.FromAnonymousUser("t2")

		ctx := context.Background()
		ctx = actor.WithActor(ctx, actor1)
		ctx = WithFlags(ctx, mockStore)

		// Make sure user1 flags are set
		require.True(t, EvaluateForActorFromContext(ctx, "f1"))
		require.False(t, EvaluateForActorFromContext(ctx, "f2"))

		// With a new actor, the flag fetcher should re-fetch
		ctx = actor.WithActor(ctx, actor2)
		require.True(t, EvaluateForActorFromContext(ctx, "f2"))
		require.False(t, EvaluateForActorFromContext(ctx, "f1"))

		mockrequire.CalledN(t, mockStore.GetAnonymousUserFlagFunc, 4)
		// TODO: test that for each above cases setEvaluatedFlagToCache has been called with proper args
	})
	// TODO: case 3: test that for rest GetGlobalFeatureFlagFunc is called
	t.Run("no user", func(t *testing.T) {})
}

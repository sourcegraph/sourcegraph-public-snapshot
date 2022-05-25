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

func TestEvaluateForActorFromContext(t *testing.T) {
	t.Run("for authenticated user", func(t *testing.T) {
		mockStore := NewMockStore()

		mockStore.GetUserFlagFunc.SetDefaultHook(func(_ context.Context, uid int32, flag string) (*bool, error) {
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
	})

	t.Run("for anonymous user", func(t *testing.T) {
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
	})

	t.Run("for no user", func(t *testing.T) {
		mockStore := NewMockStore()

		flagValue := true
		mockStore.GetGlobalFeatureFlagFunc.SetDefaultReturn(&flagValue, nil)

		ctx := context.Background()
		ctx = WithFlags(ctx, mockStore)

		require.True(t, EvaluateForActorFromContext(ctx, "test-flag"))

		mockrequire.CalledN(t, mockStore.GetGlobalFeatureFlagFunc, 1)
	})
}

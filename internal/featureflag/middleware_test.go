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
	mockStore.GetUserFlagsFunc.SetDefaultReturn(map[string]bool{"user1": true}, nil)

	handler := http.Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// After going through the middleware, a request with an actor should
		// also have feature flags available.
		require.True(t, FromContext(r.Context())["user1"])
	}))
	handler = Middleware(mockStore, handler)

	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func TestContextFlags(t *testing.T) {
	mockStore := NewMockStore()
	mockStore.GetUserFlagsFunc.SetDefaultHook(func(_ context.Context, uid int32) (map[string]bool, error) {
		switch uid {
		case 1:
			return map[string]bool{"user1": true}, nil
		case 2:
			return map[string]bool{"user2": true}, nil
		default:
			return map[string]bool{}, nil
		}
	})

	actor1 := actor.FromUser(1)
	actor2 := actor.FromUser(2)

	ctx := context.Background()
	ctx = actor.WithActor(ctx, actor1)
	ctx = WithFlags(ctx, mockStore)

	// Make sure user1 flags are set
	flags := FromContext(ctx)
	require.True(t, flags["user1"])

	// With a new actor, the flag fetcher should re-fetch
	ctx = actor.WithActor(ctx, actor2)
	flags = FromContext(ctx)
	require.False(t, flags["user1"])
	require.True(t, flags["user2"])

	// With the first actor, we should return flags for the first actor and we
	// should not call GetUserFlags again because the flags should be cached.
	ctx = actor.WithActor(ctx, actor1)
	flags = FromContext(ctx)
	require.True(t, flags["user1"])
	require.False(t, flags["user2"])
	mockrequire.CalledN(t, mockStore.GetUserFlagsFunc, 2)
}

package featureflag

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestOrgFeatureFlagOverride(t *testing.T) {
	mockStore := NewMockStore()

	mockStore.GetFeatureFlagFunc.SetDefaultHook(func(ctx context.Context, name string) (*FeatureFlag, error) {
		return &FeatureFlag{
			Name:      name,
			Bool:      &FeatureFlagBool{Value: false},
			Rollout:   nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
		}, nil
	})
	mockStore.GetUserOverrideFunc.SetDefaultReturn(nil, nil)
	mockStore.GetOrgOverrideForUserFunc.SetDefaultReturn(nil, nil)

	actor1 := actor.FromUser(1)
	ctx := context.Background()
	ctx = actor.WithActor(ctx, actor1)
	ctx = WithFlags(ctx, mockStore)

	require.False(t, EvaluateForActorFromContext(ctx, "test-flag"))

	mockStore.GetOrgOverrideForUserFunc.SetDefaultHook(func(ctx context.Context, uid int32, flag string) (*Override, error) {
		var orgID int32 = 1

		return &Override{
			UserID:   nil,
			OrgID:    &orgID,
			FlagName: flag,
			Value:    true,
		}, nil
	})

	r, _ := mockStore.GetOrgOverrideForUser(ctx, 1, "test-flag")
	require.True(t, r.Value)

	require.True(t, EvaluateForActorFromContext(ctx, "test-flag"))
}

func TestEvaluateForActorFromContext(t *testing.T) {
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

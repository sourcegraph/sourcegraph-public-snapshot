package featureflag

import (
	"context"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

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

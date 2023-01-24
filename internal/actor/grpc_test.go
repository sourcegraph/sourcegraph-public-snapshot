package actor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActorPropagator(t *testing.T) {
	t.Run("no actor", func(t *testing.T) {
		ap := ActorPropagator{}
		md := ap.FromContext(context.Background())
		ctx := ap.InjectContext(context.Background(), md)
		actor := FromContext(ctx)
		require.False(t, actor.IsAuthenticated())
	})

	t.Run("internal actor", func(t *testing.T) {
		ap := ActorPropagator{}
		ctx1 := WithInternalActor(context.Background())
		md := ap.FromContext(ctx1)
		ctx2 := ap.InjectContext(context.Background(), md)
		actor := FromContext(ctx2)
		require.True(t, actor.IsInternal())
	})

	t.Run("user actor", func(t *testing.T) {
		ap := ActorPropagator{}
		ctx1 := WithActor(context.Background(), FromUser(16))
		md := ap.FromContext(ctx1)
		ctx2 := ap.InjectContext(context.Background(), md)
		actor := FromContext(ctx2)
		require.True(t, actor.IsAuthenticated())
		require.Equal(t, int32(16), actor.UID)
	})

	t.Run("anonymous user actor", func(t *testing.T) {
		ap := ActorPropagator{}
		ctx1 := WithActor(context.Background(), FromAnonymousUser("anon123"))
		md := ap.FromContext(ctx1)
		ctx2 := ap.InjectContext(context.Background(), md)
		actor := FromContext(ctx2)
		require.Equal(t, "anon123", actor.AnonymousUID)
	})
}

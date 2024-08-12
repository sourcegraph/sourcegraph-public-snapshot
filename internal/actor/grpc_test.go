package actor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestActorPropagator(t *testing.T) {
	t.Run("no actor", func(t *testing.T) {
		ap := ActorPropagator{}
		md := ap.FromContext(context.Background())
		ctx, err := ap.InjectContext(context.Background(), md)
		require.NoError(t, err)
		actor := FromContext(ctx)
		require.False(t, actor.IsAuthenticated())
	})

	t.Run("internal actor", func(t *testing.T) {
		ap := ActorPropagator{}
		ctx1 := WithInternalActor(context.Background())
		md := ap.FromContext(ctx1)
		ctx2, err := ap.InjectContext(context.Background(), md)
		require.NoError(t, err)
		actor := FromContext(ctx2)
		require.True(t, actor.IsInternal())
	})

	t.Run("user actor", func(t *testing.T) {
		ap := ActorPropagator{}
		ctx1 := WithActor(context.Background(), FromUser(16))
		md := ap.FromContext(ctx1)
		ctx2, err := ap.InjectContext(context.Background(), md)
		require.NoError(t, err)
		actor := FromContext(ctx2)
		require.True(t, actor.IsAuthenticated())
		require.Equal(t, int32(16), actor.UID)
	})

	t.Run("bad actor value", func(t *testing.T) {
		ap := ActorPropagator{}
		md := make(metadata.MD)
		md.Append(headerKeyActorUID, "suchabadvalue")
		_, err := ap.InjectContext(context.Background(), md)
		require.Error(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("anonymous user actor", func(t *testing.T) {
		ap := ActorPropagator{}
		ctx1 := WithActor(context.Background(), FromAnonymousUser("anon123"))
		md := ap.FromContext(ctx1)
		ctx2, err := ap.InjectContext(context.Background(), md)
		require.NoError(t, err)
		actor := FromContext(ctx2)
		require.Equal(t, "anon123", actor.AnonymousUID)
	})

	t.Run("user actor with anonymous UID", func(t *testing.T) {
		originalActor := FromUser(16)
		originalActor.AnonymousUID = "foobar"
		ctx1 := WithActor(context.Background(), originalActor)

		ap := ActorPropagator{}
		md := ap.FromContext(ctx1)
		ctx2, err := ap.InjectContext(context.Background(), md)
		require.NoError(t, err)
		actor := FromContext(ctx2)
		require.True(t, actor.IsAuthenticated())
		require.Equal(t, "foobar", actor.AnonymousUID)
		require.Equal(t, int32(16), actor.UID)
	})
}

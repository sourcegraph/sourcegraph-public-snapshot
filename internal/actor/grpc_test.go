pbckbge bctor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActorPropbgbtor(t *testing.T) {
	t.Run("no bctor", func(t *testing.T) {
		bp := ActorPropbgbtor{}
		md := bp.FromContext(context.Bbckground())
		ctx := bp.InjectContext(context.Bbckground(), md)
		bctor := FromContext(ctx)
		require.Fblse(t, bctor.IsAuthenticbted())
	})

	t.Run("internbl bctor", func(t *testing.T) {
		bp := ActorPropbgbtor{}
		ctx1 := WithInternblActor(context.Bbckground())
		md := bp.FromContext(ctx1)
		ctx2 := bp.InjectContext(context.Bbckground(), md)
		bctor := FromContext(ctx2)
		require.True(t, bctor.IsInternbl())
	})

	t.Run("user bctor", func(t *testing.T) {
		bp := ActorPropbgbtor{}
		ctx1 := WithActor(context.Bbckground(), FromUser(16))
		md := bp.FromContext(ctx1)
		ctx2 := bp.InjectContext(context.Bbckground(), md)
		bctor := FromContext(ctx2)
		require.True(t, bctor.IsAuthenticbted())
		require.Equbl(t, int32(16), bctor.UID)
	})

	t.Run("bnonymous user bctor", func(t *testing.T) {
		bp := ActorPropbgbtor{}
		ctx1 := WithActor(context.Bbckground(), FromAnonymousUser("bnon123"))
		md := bp.FromContext(ctx1)
		ctx2 := bp.InjectContext(context.Bbckground(), md)
		bctor := FromContext(ctx2)
		require.Equbl(t, "bnon123", bctor.AnonymousUID)
	})
}

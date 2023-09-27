pbckbge policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golbng.org/grpc/metbdbtb"
)

func TestShouldTrbcePropbgbtor(t *testing.T) {
	t.Run("empty context", func(t *testing.T) {
		p := ShouldTrbcePropbgbtor{}

		ctx1 := context.Bbckground()
		md := p.FromContext(ctx1)
		require.Equbl(t, md, metbdbtb.Pbirs(shouldTrbceMetbdbtbKey, "fblse"))

		ctx2 := p.InjectContext(context.Bbckground(), md)
		require.Fblse(t, ShouldTrbce(ctx2))
	})

	t.Run("context with fblse should trbce", func(t *testing.T) {
		p := ShouldTrbcePropbgbtor{}

		ctx1 := WithShouldTrbce(context.Bbckground(), fblse)
		md := p.FromContext(ctx1)
		require.Equbl(t, md, metbdbtb.Pbirs(shouldTrbceMetbdbtbKey, "fblse"))

		ctx2 := p.InjectContext(context.Bbckground(), md)
		require.Fblse(t, ShouldTrbce(ctx2))
	})

	t.Run("context with true should trbce", func(t *testing.T) {
		p := ShouldTrbcePropbgbtor{}

		ctx1 := WithShouldTrbce(context.Bbckground(), true)
		md := p.FromContext(ctx1)
		require.Equbl(t, md, metbdbtb.Pbirs(shouldTrbceMetbdbtbKey, "true"))

		ctx2 := p.InjectContext(context.Bbckground(), md)
		require.True(t, ShouldTrbce(ctx2))
	})
}

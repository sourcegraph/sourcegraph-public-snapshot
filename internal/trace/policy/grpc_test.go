package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestShouldTracePropagator(t *testing.T) {
	t.Run("empty context", func(t *testing.T) {
		p := ShouldTracePropagator{}

		ctx1 := context.Background()
		md := p.FromContext(ctx1)
		require.Equal(t, md, metadata.Pairs(shouldTraceMetadataKey, "false"))

		ctx2, err := p.InjectContext(context.Background(), md)
		require.NoError(t, err)
		require.False(t, ShouldTrace(ctx2))
	})

	t.Run("context with false should trace", func(t *testing.T) {
		p := ShouldTracePropagator{}

		ctx1 := WithShouldTrace(context.Background(), false)
		md := p.FromContext(ctx1)
		require.Equal(t, md, metadata.Pairs(shouldTraceMetadataKey, "false"))

		ctx2, err := p.InjectContext(context.Background(), md)
		require.NoError(t, err)
		require.False(t, ShouldTrace(ctx2))
	})

	t.Run("context with true should trace", func(t *testing.T) {
		p := ShouldTracePropagator{}

		ctx1 := WithShouldTrace(context.Background(), true)
		md := p.FromContext(ctx1)
		require.Equal(t, md, metadata.Pairs(shouldTraceMetadataKey, "true"))

		ctx2, err := p.InjectContext(context.Background(), md)
		require.NoError(t, err)
		require.True(t, ShouldTrace(ctx2))
	})
}

package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTenantPropagator(t *testing.T) {
	t.Run("no tenant", func(t *testing.T) {
		tp := TenantPropagator{}
		md := tp.FromContext(context.Background())
		ctx, err := tp.InjectContext(context.Background(), md)
		require.NoError(t, err)
		_, ok := FromContext(ctx)
		require.False(t, ok)
	})

	t.Run("with tenant", func(t *testing.T) {
		tp := TenantPropagator{}
		tenantID := 1
		ctx1 := withTenant(context.Background(), tenantID)
		md := tp.FromContext(ctx1)
		ctx2, err := tp.InjectContext(context.Background(), md)
		require.NoError(t, err)
		tenant, ok := FromContext(ctx2)
		require.True(t, ok)
		require.Equal(t, tenantID, tenant.ID())
	})

	t.Run("bad tenant value", func(t *testing.T) {
		tp := TenantPropagator{}
		md := make(metadata.MD)
		md.Append(headerKeyTenantID, "suchabadvalue")
		_, err := tp.InjectContext(context.Background(), md)
		require.Error(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, s.Code())
	})
}

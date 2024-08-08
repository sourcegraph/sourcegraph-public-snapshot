package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromContext(t *testing.T) {
	t.Run("no tenant", func(t *testing.T) {
		ctx := context.Background()
		_, err := FromContext(ctx)
		require.Error(t, err)
	})

	t.Run("with tenant", func(t *testing.T) {
		ctx := withTenant(context.Background(), 42)
		tenant, err := FromContext(ctx)
		require.NoError(t, err)
		require.Equal(t, &Tenant{_id: 42}, tenant)
	})
}

func TestInherit(t *testing.T) {
	t.Run("inherit from empty to empty", func(t *testing.T) {
		from := context.Background()
		to := context.Background()
		ctx, err := Inherit(from, to)
		require.NoError(t, err)
		_, err = FromContext(ctx)
		require.Error(t, err)
	})

	t.Run("inherit from tenant to empty", func(t *testing.T) {
		from := withTenant(context.Background(), 42)
		to := context.Background()
		ctx, err := Inherit(from, to)
		require.NoError(t, err)
		tenant, err := FromContext(ctx)
		require.NoError(t, err)
		require.Equal(t, 42, tenant.ID())
	})

	t.Run("inherit to existing tenant that matches from", func(t *testing.T) {
		from := withTenant(context.Background(), 42)
		to := withTenant(context.Background(), 42)
		ctx, err := Inherit(from, to)
		require.NoError(t, err)
		tenant, err := FromContext(ctx)
		require.NoError(t, err)
		require.Equal(t, 42, tenant.ID())
	})

	t.Run("inherit to different tenant", func(t *testing.T) {
		from := withTenant(context.Background(), 42)
		to := withTenant(context.Background(), 24)
		_, err := Inherit(from, to)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot inherit into a context with a tenant")
	})
}

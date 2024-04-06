package cody

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestIsCodyEnabled(t *testing.T) {
	oldMock := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(feature licensing.Feature) error {
		return nil
	}
	t.Cleanup(func() {
		licensing.MockCheckFeature = oldMock
	})

	truePtr := true
	falsePtr := false

	type Perm struct {
		namespace rtypes.PermissionNamespace
		action    rtypes.NamespaceAction
	}
	mockDB := func(perms map[Perm]bool) *dbmocks.MockDB {
		db := dbmocks.NewMockDB()
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
			return &types.User{ID: 1, SiteAdmin: false}, nil
		})
		db.UsersFunc.SetDefaultReturn(users)
		permissions := dbmocks.NewMockPermissionStore()
		permissions.GetPermissionForUserFunc.SetDefaultHook(func(ctx context.Context, opt database.GetPermissionForUserOpts) (*types.Permission, error) {
			if hasPermission, ok := perms[Perm{opt.Namespace, opt.Action}]; ok && hasPermission {
				return &types.Permission{ID: 1, Namespace: opt.Namespace, Action: opt.Action, CreatedAt: time.Now()}, nil
			}
			return nil, nil
		})
		db.PermissionsFunc.SetDefaultReturn(permissions)
		return db
	}
	defaultUserPerms := map[Perm]bool{
		{rtypes.CodyNamespace, rtypes.CodyAccessAction}: true, // Cody access
	}

	t.Run("no RBAC, Unauthenticated user", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled:     &truePtr,
				CodyPermissions: &falsePtr, // no RBAC
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false for unauthenticated actor")
		require.Equal(t, "not authenticated", reason)
	})

	t.Run("disabled by license", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})

		oldMock := licensing.MockCheckFeature
		licensing.MockCheckFeature = func(feature licensing.Feature) error {
			return licensing.NewFeatureNotActivatedError("no cody for you")
		}
		t.Cleanup(func() {
			licensing.MockCheckFeature = oldMock
		})

		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false for authenticated actor with license disabling cody")
		require.Equal(t, "instance license does not allow cody", reason)
	})

	t.Run("no RBAC, Authenticated user", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled:     &truePtr,
				CodyPermissions: &falsePtr, // no RBAC
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.True(t, enabled, "Expected IsCodyEnabled to return true for authenticated actor")
		require.Equal(t, "", reason)
	})

	t.Run("no RBAC, Enabled cody, but not completions", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled:     &truePtr,
				CodyPermissions: &falsePtr, // no RBAC
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.True(t, enabled, "Expected IsCodyEnabled to return true without completions")
		require.Equal(t, "", reason)
	})

	t.Run("no RBAC, Disabled cody", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled:     &falsePtr,
				CodyPermissions: &falsePtr, // no RBAC
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false when cody is disabled")
		require.Equal(t, "cody is disabled", reason)
	})

	t.Run("no RBAC, No cody config, default value", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyPermissions: &falsePtr, // no RBAC
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false when cody is not configured")
		require.Equal(t, "cody is disabled", reason)
	})

	t.Run("no RBAC, Cody.RestrictUsersFeatureFlag", func(t *testing.T) {
		t.Run("feature flag disabled", func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					CodyEnabled:                  &truePtr,
					CodyRestrictUsersFeatureFlag: &truePtr,
					CodyPermissions:              &falsePtr, // no RBAC
				},
			})
			t.Cleanup(func() {
				conf.Mock(nil)
			})

			db := mockDB(defaultUserPerms)
			ctx := context.Background()
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
			enabled, reason := IsCodyEnabled(ctx, db)
			require.False(t, enabled, "Expected IsCodyEnabled to return false for unauthenticated user with cody.restrictUsersFeatureFlag enabled")
			require.Equal(t, "not authenticated", reason)

			ctx = context.Background()
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
			enabled, reason = IsCodyEnabled(ctx, db)
			require.False(t, enabled, "Expected IsCodyEnabled to return false for authenticated user when cody.restrictUsersFeatureFlag is set and no feature flag is present for the user")
			require.Equal(t, "cody is restricted to feature flag but feature flag is not enabled", reason)
		})
		t.Run("feature flag enabled", func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					CodyEnabled:                  &truePtr,
					CodyRestrictUsersFeatureFlag: &truePtr,
					CodyPermissions:              &falsePtr, // no RBAC
				},
			})
			t.Cleanup(func() {
				conf.Mock(nil)
			})

			db := mockDB(defaultUserPerms)
			ctx := context.Background()
			ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"cody": true}, map[string]bool{"cody": true}, nil))
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
			enabled, reason := IsCodyEnabled(ctx, db)
			require.False(t, enabled, "Expected IsCodyEnabled to return false when cody feature flag is enabled")
			require.Equal(t, "not authenticated", reason)

			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
			enabled, reason = IsCodyEnabled(ctx, db)
			require.True(t, enabled, "Expected IsCodyEnabled to return true when cody feature flag is enabled")
			require.Equal(t, "", reason)
		})
	})

	t.Run("RBAC, Unauthenticated user", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				// Note: default when CodyRestrictUsersFeatureFlag and CodyPermissions are not set is
				// that CodyRestrictUsersFeatureFlag=false and (RBAC) CodyPermissions=true
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 0})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false for unauthenticated actor")
		require.Equal(t, "not authenticated", reason)
	})

	t.Run("RBAC, Authenticated user", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled: &truePtr,
				// Note: default when CodyRestrictUsersFeatureFlag and CodyPermissions are not set is
				// that CodyRestrictUsersFeatureFlag=false and (RBAC) CodyPermissions=true
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.True(t, enabled, "Expected IsCodyEnabled to return true for authenticated actor")
		require.Equal(t, "", reason)
	})

	t.Run("RBAC, Disabled cody", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled: &falsePtr,
				// Note: default when CodyRestrictUsersFeatureFlag and CodyPermissions are not set is
				// that CodyRestrictUsersFeatureFlag=false and (RBAC) CodyPermissions=true
			},
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false when cody is disabled")
		require.Equal(t, "cody is disabled", reason)
	})

	t.Run("RBAC, No cody config, default value", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{},
			// Note: default when CodyRestrictUsersFeatureFlag and CodyPermissions are not set is
			// that CodyRestrictUsersFeatureFlag=false and (RBAC) CodyPermissions=true
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(defaultUserPerms)
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false when cody is not configured")
		require.Equal(t, "cody is disabled", reason)
	})

	t.Run("RBAC, No cody permissions", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{},
			// Note: default when CodyRestrictUsersFeatureFlag and CodyPermissions are not set is
			// that CodyRestrictUsersFeatureFlag=false and (RBAC) CodyPermissions=true
		})
		t.Cleanup(func() {
			conf.Mock(nil)
		})
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
		db := mockDB(nil) // Cody access permission not granted
		enabled, reason := IsCodyEnabled(ctx, db)
		require.False(t, enabled, "Expected IsCodyEnabled to return false when user does not have cody access permission")
		require.Equal(t, "cody is disabled", reason)
	})
}

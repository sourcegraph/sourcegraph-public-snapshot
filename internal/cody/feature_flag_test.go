package cody

import (
	"context"
	"testing"
	"time"

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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false for unauthenticated actor")
		}
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
		if !IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return true for authenticated actor")
		}
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
		if !IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return true without completions")
		}
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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false when cody is disabled")
		}
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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false when cody is not configured")
		}
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
			if IsCodyEnabled(ctx, db) {
				t.Error("Expected IsCodyEnabled to return false for unauthenticated user with cody.restrictUsersFeatureFlag enabled")
			}
			ctx = context.Background()
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
			if IsCodyEnabled(ctx, db) {
				t.Error("Expected IsCodyEnabled to return false for authenticated user when cody.restrictUsersFeatureFlag is set and no feature flag is present for the user")
			}
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
			if IsCodyEnabled(ctx, db) {
				t.Error("Expected IsCodyEnabled to return false when cody feature flag is enabled")
			}
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
			if !IsCodyEnabled(ctx, db) {
				t.Error("Expected IsCodyEnabled to return true when cody feature flag is enabled")
			}
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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false for unauthenticated actor")
		}
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
		if !IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return true for authenticated actor")
		}
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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false when cody is disabled")
		}
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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false when cody is not configured")
		}
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
		if IsCodyEnabled(ctx, db) {
			t.Error("Expected IsCodyEnabled to return false when user does not have cody access permission")
		}
	})
}

package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagStore(t *testing.T) {
	t.Parallel()
	t.Run("NewFeatureFlag", testNewFeatureFlagRoundtrip)
	t.Run("ListFeatureFlags", testListFeatureFlags)
	t.Run("Overrides", func(t *testing.T) {
		t.Run("NewOverride", testNewOverrideRoundtrip)
		t.Run("ListUserOverrides", testListUserOverrides)
		t.Run("ListOrgOverrides", testListOrgOverrides)
	})
	t.Run("UserFlags", testUserFlags)
	t.Run("AnonymousUserFlags", testAnonymousUserFlags)
	t.Run("UserlessFeatureFlags", testUserlessFeatureFlags)
}

func errorContains(s string) require.ErrorAssertionFunc {
	return func(t require.TestingT, err error, msg ...interface{}) {
		require.Error(t, err)
		require.Contains(t, err.Error(), s, msg)
	}
}

func cleanup(t *testing.T, db *sql.DB) func() {
	return func() {
		if t.Failed() {
			// Retain content on failed tests
			return
		}
		_, err := db.Exec(`truncate feature_flags, feature_flag_overrides, users, orgs, org_members cascade;`)
		require.NoError(t, err)
	}
}

func testNewFeatureFlagRoundtrip(t *testing.T) {
	t.Parallel()
	ff := FeatureFlags(dbtest.NewDB(t, ""))
	ctx := actor.WithInternalActor(context.Background())

	cases := []struct {
		flag      *types.FeatureFlag
		assertErr require.ErrorAssertionFunc
	}{
		{
			flag: &types.FeatureFlag{Name: "bool_true", Bool: &types.FeatureFlagBool{Value: true}},
		},
		{
			flag: &types.FeatureFlag{Name: "bool_false", Bool: &types.FeatureFlagBool{Value: false}},
		},
		{
			flag: &types.FeatureFlag{Name: "min_rollout", BoolVar: &types.FeatureFlagBoolVar{Rollout: 0}},
		},
		{
			flag: &types.FeatureFlag{Name: "mid_rollout", BoolVar: &types.FeatureFlagBoolVar{Rollout: 3124}},
		},
		{
			flag: &types.FeatureFlag{Name: "max_rollout", BoolVar: &types.FeatureFlagBoolVar{Rollout: 10000}},
		},
		{
			flag:      &types.FeatureFlag{Name: "err_too_high_rollout", BoolVar: &types.FeatureFlagBoolVar{Rollout: 10001}},
			assertErr: errorContains(`violates check constraint "feature_flags_rollout_check"`),
		},
		{
			flag:      &types.FeatureFlag{Name: "err_too_low_rollout", BoolVar: &types.FeatureFlagBoolVar{Rollout: -1}},
			assertErr: errorContains(`violates check constraint "feature_flags_rollout_check"`),
		},
		{
			flag:      &types.FeatureFlag{Name: "err_no_types"},
			assertErr: errorContains(`feature flag must have exactly one type`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.flag.Name, func(t *testing.T) {
			res, err := ff.NewFeatureFlag(ctx, tc.flag)
			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}
			require.NoError(t, err)

			// Only assert that the values it is created with are equal.
			// Don't bother with the timestamps
			require.Equal(t, tc.flag.Name, res.Name)
			require.Equal(t, tc.flag.Bool, res.Bool)
			require.Equal(t, tc.flag.BoolVar, res.BoolVar)
		})
	}
}

func testListFeatureFlags(t *testing.T) {
	t.Parallel()
	ff := FeatureFlags(dbtest.NewDB(t, ""))
	ctx := actor.WithInternalActor(context.Background())

	flag1 := &types.FeatureFlag{Name: "bool_true", Bool: &types.FeatureFlagBool{Value: true}}
	flag2 := &types.FeatureFlag{Name: "bool_false", Bool: &types.FeatureFlagBool{Value: false}}
	flag3 := &types.FeatureFlag{Name: "mid_rollout", BoolVar: &types.FeatureFlagBoolVar{Rollout: 3124}}
	flags := []*types.FeatureFlag{flag1, flag2, flag3}

	for _, flag := range flags {
		_, err := ff.NewFeatureFlag(ctx, flag)
		require.NoError(t, err)
	}

	res, err := ff.ListFeatureFlags(ctx)
	require.NoError(t, err)
	for _, flag := range res {
		// Unset any timestamps
		flag.CreatedAt = time.Time{}
		flag.UpdatedAt = time.Time{}
		flag.DeletedAt = nil
	}

	require.EqualValues(t, res, flags)
}

func testNewOverrideRoundtrip(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ff := FeatureFlags(db)
	users := Users(db)
	ctx := actor.WithInternalActor(context.Background())

	ff1, err := ff.NewBool(ctx, "t", true)
	require.NoError(t, err)

	u1, err := users.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	invalidUserID := int32(38535)

	cases := []struct {
		override  *types.FeatureFlagOverride
		assertErr require.ErrorAssertionFunc
	}{
		{
			override: &types.FeatureFlagOverride{UserID: &u1.ID, FlagName: ff1.Name, Value: false},
		},
		{
			override:  &types.FeatureFlagOverride{UserID: &invalidUserID, FlagName: ff1.Name, Value: false},
			assertErr: errorContains(`violates foreign key constraint "feature_flag_overrides_namespace_user_id_fkey"`),
		},
		{
			override:  &types.FeatureFlagOverride{UserID: &u1.ID, FlagName: "invalid-flag-name", Value: false},
			assertErr: errorContains(`violates foreign key constraint "feature_flag_overrides_flag_name_fkey"`),
		},
		{
			override:  &types.FeatureFlagOverride{FlagName: ff1.Name, Value: false},
			assertErr: errorContains(`violates check constraint "feature_flag_overrides_has_org_or_user_id"`),
		},
	}

	for _, tc := range cases {
		t.Run("case", func(t *testing.T) {
			res, err := ff.NewOverride(ctx, tc.override)
			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.override, res)
		})
	}
}

func testListUserOverrides(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ff := FeatureFlags(db)
	users := Users(db)
	ctx := actor.WithInternalActor(context.Background())

	mkUser := func(name string) *types.User {
		u, err := users.Create(ctx, NewUser{Username: name, Password: "p"})
		require.NoError(t, err)
		return u
	}

	mkFFBool := func(name string, val bool) *types.FeatureFlag {
		ff, err := ff.NewBool(ctx, name, val)
		require.NoError(t, err)
		return ff
	}

	mkOverride := func(user int32, flag string, val bool) *types.FeatureFlagOverride {
		ffo, err := ff.NewOverride(ctx, &types.FeatureFlagOverride{UserID: &user, FlagName: flag, Value: val})
		require.NoError(t, err)
		return ffo
	}

	t.Run("no overrides", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		mkFFBool("f", true)
		got, err := ff.ListUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("some overrides", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		f1 := mkFFBool("f", true)
		o1 := mkOverride(u1.ID, f1.Name, false)
		got, err := ff.ListUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Equal(t, got, []*types.FeatureFlagOverride{o1})
	})

	t.Run("overrides for other users", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u1")
		u2 := mkUser("u2")
		f1 := mkFFBool("f", true)
		o1 := mkOverride(u1.ID, f1.Name, false)
		mkOverride(u2.ID, f1.Name, true)
		got, err := ff.ListUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Equal(t, got, []*types.FeatureFlagOverride{o1})
	})

	t.Run("non-unique override errors", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u1")
		f1 := mkFFBool("f", true)
		_, err := ff.NewOverride(ctx, &types.FeatureFlagOverride{UserID: &u1.ID, FlagName: f1.Name, Value: true})
		require.NoError(t, err)
		_, err = ff.NewOverride(ctx, &types.FeatureFlagOverride{UserID: &u1.ID, FlagName: f1.Name, Value: true})
		require.Error(t, err)
	})
}

func testListOrgOverrides(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ff := FeatureFlags(db)
	users := Users(db)
	orgs := Orgs(db)
	orgMembers := OrgMembers(db)
	ctx := actor.WithInternalActor(context.Background())

	mkUser := func(name string, orgIDs ...int32) *types.User {
		u, err := users.Create(ctx, NewUser{Username: name, Password: "p"})
		require.NoError(t, err)
		for _, id := range orgIDs {
			_, err := orgMembers.Create(ctx, id, u.ID)
			require.NoError(t, err)
		}
		return u
	}

	mkFFBool := func(name string, val bool) *types.FeatureFlag {
		ff, err := ff.NewBool(ctx, name, val)
		require.NoError(t, err)
		return ff
	}

	mkOverride := func(org int32, flag string, val bool) *types.FeatureFlagOverride {
		ffo, err := ff.NewOverride(ctx, &types.FeatureFlagOverride{OrgID: &org, FlagName: flag, Value: val})
		require.NoError(t, err)
		return ffo
	}

	mkOrg := func(name string) *types.Org {
		o, err := orgs.Create(ctx, name, nil)
		require.NoError(t, err)
		return o
	}

	t.Run("no overrides", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		mkFFBool("f", true)
		got, err := ff.ListUserOverrides(ctx, u1.ID)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("some overrides", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		org1 := mkOrg("org1")
		u1 := mkUser("u", org1.ID)
		f1 := mkFFBool("f", true)
		o1 := mkOverride(org1.ID, f1.Name, false)
		got, err := ff.ListOrgOverridesForUser(ctx, u1.ID)
		require.NoError(t, err)
		require.Equal(t, got, []*types.FeatureFlagOverride{o1})
	})

	t.Run("non-unique override errors", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		org1 := mkOrg("org1")
		f1 := mkFFBool("f", true)

		_, err := ff.NewOverride(ctx, &types.FeatureFlagOverride{OrgID: &org1.ID, FlagName: f1.Name, Value: true})
		require.NoError(t, err)
		_, err = ff.NewOverride(ctx, &types.FeatureFlagOverride{OrgID: &org1.ID, FlagName: f1.Name, Value: false})
		require.Error(t, err)
	})
}

func testUserFlags(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ff := FeatureFlags(db)
	users := Users(db)
	orgs := Orgs(db)
	orgMembers := OrgMembers(db)
	ctx := actor.WithInternalActor(context.Background())

	mkUser := func(name string, orgIDs ...int32) *types.User {
		u, err := users.Create(ctx, NewUser{Username: name, Password: "p"})
		require.NoError(t, err)
		for _, id := range orgIDs {
			_, err := orgMembers.Create(ctx, id, u.ID)
			require.NoError(t, err)
		}
		return u
	}

	mkFFBool := func(name string, val bool) *types.FeatureFlag {
		ff, err := ff.NewBool(ctx, name, val)
		require.NoError(t, err)
		return ff
	}

	mkFFBoolVar := func(name string, rollout int) *types.FeatureFlag {
		ff, err := ff.NewBoolVar(ctx, name, rollout)
		require.NoError(t, err)
		return ff
	}

	mkUserOverride := func(user int32, flag string, val bool) *types.FeatureFlagOverride {
		ffo, err := ff.NewOverride(ctx, &types.FeatureFlagOverride{UserID: &user, FlagName: flag, Value: val})
		require.NoError(t, err)
		return ffo
	}

	mkOrgOverride := func(org int32, flag string, val bool) *types.FeatureFlagOverride {
		ffo, err := ff.NewOverride(ctx, &types.FeatureFlagOverride{OrgID: &org, FlagName: flag, Value: val})
		require.NoError(t, err)
		return ffo
	}

	mkOrg := func(name string) *types.Org {
		o, err := orgs.Create(ctx, name, nil)
		require.NoError(t, err)
		return o
	}

	t.Run("bool vals", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		mkFFBool("f1", true)
		mkFFBool("f2", false)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": false}
		require.Equal(t, expected, got)
	})

	t.Run("bool vars", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		mkFFBoolVar("f1", 10000)
		mkFFBoolVar("f2", 0)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": false}
		require.Equal(t, expected, got)
	})

	t.Run("bool vals with user override", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		mkFFBool("f1", true)
		mkFFBool("f2", false)
		mkUserOverride(u1.ID, "f2", true)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": true}
		require.Equal(t, expected, got)
	})

	t.Run("bool vars with user override", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		u1 := mkUser("u")
		mkFFBoolVar("f1", 10000)
		mkFFBoolVar("f2", 0)
		mkUserOverride(u1.ID, "f2", true)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": true}
		require.Equal(t, expected, got)
	})

	t.Run("bool vals with org override", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		mkFFBool("f1", true)
		mkFFBool("f2", false)
		mkOrgOverride(o1.ID, "f2", true)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": true}
		require.Equal(t, expected, got)
	})

	t.Run("bool vars with org override", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		mkFFBoolVar("f1", 10000)
		mkFFBoolVar("f2", 0)
		mkOrgOverride(o1.ID, "f2", true)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": true}
		require.Equal(t, expected, got)
	})

	t.Run("user override beats org override", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		o1 := mkOrg("o1")
		u1 := mkUser("u", o1.ID)
		mkFFBoolVar("f1", 10000)
		mkFFBoolVar("f2", 0)
		mkOrgOverride(o1.ID, "f2", true)
		mkUserOverride(u1.ID, "f2", false)

		got, err := ff.UserFlags(ctx, u1.ID)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": false}
		require.Equal(t, expected, got)
	})
}

func testAnonymousUserFlags(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ff := FeatureFlags(db)
	ctx := actor.WithInternalActor(context.Background())

	mkFFBool := func(name string, val bool) *types.FeatureFlag {
		ff, err := ff.NewBool(ctx, name, val)
		require.NoError(t, err)
		return ff
	}

	mkFFBoolVar := func(name string, rollout int) *types.FeatureFlag {
		ff, err := ff.NewBoolVar(ctx, name, rollout)
		require.NoError(t, err)
		return ff
	}

	t.Run("bool vals", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		mkFFBool("f1", true)
		mkFFBool("f2", false)

		got, err := ff.AnonymousUserFlags(ctx, "testuser")
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": false}
		require.Equal(t, expected, got)
	})

	t.Run("bool vars", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		mkFFBoolVar("f1", 10000)
		mkFFBoolVar("f2", 0)

		got, err := ff.AnonymousUserFlags(ctx, "testuser")
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": false}
		require.Equal(t, expected, got)
	})

	// No override tests for AnonymousUserFlags because no override
	// can be defined for an anonymous user.
}

func testUserlessFeatureFlags(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ff := FeatureFlags(db)
	ctx := actor.WithInternalActor(context.Background())

	mkFFBool := func(name string, val bool) *types.FeatureFlag {
		ff, err := ff.NewBool(ctx, name, val)
		require.NoError(t, err)
		return ff
	}

	mkFFBoolVar := func(name string, rollout int) *types.FeatureFlag {
		ff, err := ff.NewBoolVar(ctx, name, rollout)
		require.NoError(t, err)
		return ff
	}

	t.Run("bool vals", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		mkFFBool("f1", true)
		mkFFBool("f2", false)

		got, err := ff.UserlessFeatureFlags(ctx)
		require.NoError(t, err)
		expected := map[string]bool{"f1": true, "f2": false}
		require.Equal(t, expected, got)
	})

	t.Run("bool vars", func(t *testing.T) {
		t.Cleanup(cleanup(t, db))
		mkFFBoolVar("f1", 10000)
		mkFFBoolVar("f2", 0)

		got, err := ff.UserlessFeatureFlags(ctx)
		require.NoError(t, err)

		// Userless requests don't have a stable user to evaluate
		// bool variable flags, so none should be defined.
		//
		// TODO(camdencheek): consider evaluating rollout feature
		// flags with a static string so they are defined and stable,
		// but effectively statically random.
		expected := map[string]bool{}
		require.Equal(t, expected, got)
	})
}

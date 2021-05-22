package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagStore(t *testing.T) {
	t.Run("NewFeatureFlag", testNewFeatureFlagRoundtrip)
	t.Run("ListFeatureFlags", testListFeatureFlags)
	t.Run("Overrides", func(t *testing.T) {
		t.Run("NewOverride", testNewOverrideRoundtrip)
	})
}

func errorContains(s string) require.ErrorAssertionFunc {
	return func(t require.TestingT, err error, msg ...interface{}) {
		require.Error(t, err)
		require.Contains(t, err.Error(), s, msg)
	}
}

func testNewFeatureFlagRoundtrip(t *testing.T) {
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

			// Only assert that the values it is created with are equal.
			// Don't bother with the timestamps
			require.Equal(t, tc.flag.Name, res.Name)
			require.Equal(t, tc.flag.Bool, res.Bool)
			require.Equal(t, tc.flag.BoolVar, res.BoolVar)
		})
	}
}

func testListFeatureFlags(t *testing.T) {
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
	ff := FeatureFlags(dbtest.NewDB(t, ""))
	ctx := actor.WithInternalActor(context.Background())

	cases := []struct {
		override  *types.FeatureFlagOverride
		assertErr require.ErrorAssertionFunc
	}{
		{
			override: &types.FeatureFlagOverride{},
		},
	}

	for _, tc := range cases {
		t.Run("case", func(t *testing.T) {
			res, err := ff.NewOverride(ctx, tc.override)
			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

			require.Equal(t, tc.override, res)
		})
	}
}

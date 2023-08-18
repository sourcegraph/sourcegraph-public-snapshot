package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrganizationFeatureFlagOverrides(t *testing.T) {
	t.Run("return org flag override for user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Name: "acme"}
		orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefaultReturn(&mockedOrg, nil)

		flags := dbmocks.NewMockFeatureFlagStore()
		mockedFeatureFlag := featureflag.FeatureFlag{Name: "test-flag", Bool: &featureflag.FeatureFlagBool{Value: false}, Rollout: nil, CreatedAt: time.Now(), UpdatedAt: time.Now(), DeletedAt: nil}
		mockedOverride := featureflag.Override{UserID: nil, OrgID: &mockedOrg.ID, FlagName: "test-flag", Value: true}
		flagOverrides := []*featureflag.Override{&mockedOverride}

		flags.GetFeatureFlagFunc.SetDefaultHook(func(ctx context.Context, flagName string) (*featureflag.FeatureFlag, error) {
			return &mockedFeatureFlag, nil
		})

		flags.GetOrgOverridesForUserFunc.SetDefaultHook(func(ctx context.Context, userID int32) ([]*featureflag.Override, error) {
			assert.Equal(t, int32(1), userID)
			return flagOverrides, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefaultReturn(orgs)
		db.UsersFunc.SetDefaultReturn(users)
		db.FeatureFlagsFunc.SetDefaultReturn(flags)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
				{
					organizationFeatureFlagOverrides {
						namespace {
							id
						},
						targetFlag {
							... on FeatureFlagBoolean {
								name
							},
							... on FeatureFlagRollout {
								name
							}
						},
						value
					}
				}
				`,
				ExpectedResult: `
					{
						"organizationFeatureFlagOverrides": [
							{
								"namespace": {
									"id": "T3JnOjE="
								},
								"targetFlag": {
									"name": "test-flag"
								},
								"value": true
							}
						]
					}
				`,
			},
		})
	})

	t.Run("return empty list if no overrides", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Name: "acme"}
		orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefaultReturn(&mockedOrg, nil)

		flags := dbmocks.NewMockFeatureFlagStore()
		mockedFeatureFlag := featureflag.FeatureFlag{Name: "test-flag", Bool: &featureflag.FeatureFlagBool{Value: false}, Rollout: nil, CreatedAt: time.Now(), UpdatedAt: time.Now(), DeletedAt: nil}

		flags.GetFeatureFlagFunc.SetDefaultHook(func(ctx context.Context, flagName string) (*featureflag.FeatureFlag, error) {
			return &mockedFeatureFlag, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefaultReturn(orgs)
		db.UsersFunc.SetDefaultReturn(users)
		db.FeatureFlagsFunc.SetDefaultReturn(flags)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
				{
					organizationFeatureFlagOverrides {
						namespace {
							id
						},
						targetFlag {
							... on FeatureFlagBoolean {
								name
							},
							... on FeatureFlagRollout {
								name
							}
						},
						value
					}
				}
				`,
				ExpectedResult: `
					{
						"organizationFeatureFlagOverrides": []
					}
				`,
			},
		})
	})

	t.Run("return multiple org overrides for user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Name: "acme"}
		orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefaultReturn(&mockedOrg, nil)

		flags := dbmocks.NewMockFeatureFlagStore()
		mockedFeatureFlag1 := featureflag.FeatureFlag{Name: "test-flag", Bool: &featureflag.FeatureFlagBool{Value: false}, Rollout: nil, CreatedAt: time.Now(), UpdatedAt: time.Now(), DeletedAt: nil}
		mockedFeatureFlag2 := featureflag.FeatureFlag{Name: "another-flag", Bool: &featureflag.FeatureFlagBool{Value: false}, Rollout: nil, CreatedAt: time.Now(), UpdatedAt: time.Now(), DeletedAt: nil}
		mockedOverride1 := featureflag.Override{UserID: nil, OrgID: &mockedOrg.ID, FlagName: "test-flag", Value: true}
		mockedOverride2 := featureflag.Override{UserID: nil, OrgID: &mockedOrg.ID, FlagName: "another-flag", Value: true}
		flagOverrides := []*featureflag.Override{&mockedOverride1, &mockedOverride2}

		flags.GetFeatureFlagFunc.SetDefaultHook(func(ctx context.Context, flagName string) (*featureflag.FeatureFlag, error) {
			if flagName == "test-flag" {
				return &mockedFeatureFlag1, nil
			} else {
				return &mockedFeatureFlag2, nil
			}
		})

		flags.GetOrgOverridesForUserFunc.SetDefaultHook(func(ctx context.Context, userID int32) ([]*featureflag.Override, error) {
			assert.Equal(t, int32(1), userID)
			return flagOverrides, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefaultReturn(orgs)
		db.UsersFunc.SetDefaultReturn(users)
		db.FeatureFlagsFunc.SetDefaultReturn(flags)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
				{
					organizationFeatureFlagOverrides {
						namespace {
							id
						},
						targetFlag {
							... on FeatureFlagBoolean {
								name
							},
							... on FeatureFlagRollout {
								name
							}
						},
						value
					}
				}
				`,
				ExpectedResult: `
					{
						"organizationFeatureFlagOverrides": [
							{
								"namespace": {
									"id": "T3JnOjE="
								},
								"targetFlag": {
									"name": "test-flag"
								},
								"value": true
							},
							{
								"namespace": {
									"id": "T3JnOjE="
								},
								"targetFlag": {
									"name": "another-flag"
								},
								"value": true
							}
						]
					}
				`,
			},
		})
	})
}

func TestEvaluateFeatureFlag(t *testing.T) {
	t.Run("return flag value for user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		orgs := dbmocks.NewMockOrgStore()
		mockedOrg := types.Org{ID: 1, Name: "acme"}
		orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)
		orgs.GetByIDFunc.SetDefaultReturn(&mockedOrg, nil)

		flags := dbmocks.NewMockFeatureFlagStore()
		flags.GetUserFlagsFunc.SetDefaultHook(func(ctx context.Context, uid int32) (map[string]bool, error) {
			return map[string]bool{"enabled-flag": true, "disabled-flag": false}, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefaultReturn(orgs)
		db.UsersFunc.SetDefaultReturn(users)
		db.FeatureFlagsFunc.SetDefaultReturn(flags)
		ctx = featureflag.WithFlags(ctx, flags)

		RunTests(t, []*Test{
			{
				Context: ctx,
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
				{
					evaluateFeatureFlag(flagName: "enabled-flag")
				}
				`,
				ExpectedResult: `
					{
						"evaluateFeatureFlag": true
					}
				`,
			},
			{
				Context: ctx,
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
				{
					evaluateFeatureFlag(flagName: "disabled-flag")
				}
				`,
				ExpectedResult: `
					{
						"evaluateFeatureFlag": false
					}
				`,
			},
			{
				Context: ctx,
				Schema:  mustParseGraphQLSchema(t, db),
				Query: `
				{
					evaluateFeatureFlag(flagName: "non-existing-flag")
				}
				`,
				ExpectedResult: `
					{
						"evaluateFeatureFlag": null
					}
				`,
			},
		})
	})
}

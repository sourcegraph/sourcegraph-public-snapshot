package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrganizationFeatureFlagOverrides(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	orgs := database.NewMockOrgStore()
	mockedOrg := types.Org{ID: 1, Name: "acme"}
	orgs.GetByNameFunc.SetDefaultReturn(&mockedOrg, nil)
	orgs.GetByIDFunc.SetDefaultReturn(&mockedOrg, nil)

	flags := database.NewMockFeatureFlagStore()
	mockedFlag := featureflag.Override{UserID: nil, OrgID: &mockedOrg.ID, FlagName: "test-flag", Value: true}
	flagOverrides := []*featureflag.Override{&mockedFlag}

	flags.GetOrgOverridesForUserFunc.SetDefaultHook(func(ctx context.Context, userID int32) ([]*featureflag.Override, error) {
		assert.Equal(t, int32(1), userID)
		return flagOverrides, nil
	})

	db := database.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.FeatureFlagsFunc.SetDefaultReturn(flags)

	result, err := newSchemaResolver(db).OrganizationFeatureFlagOverrides(ctx)

	if err != nil {
		t.Errorf("expected error to be nil")
	}

	got := []*featureflag.Override{}

	for _, f := range result {
		got = append(got, f.inner)
	}

	if !reflect.DeepEqual(got, flagOverrides) {
		t.Errorf("expected %v got %v", flagOverrides, got)
	}

	t.Run("return org override for user", func(t *testing.T) {
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
						"organizationFeatureFlagOverrides": {
							"nodes": [
								{
									"namespace": {
										"id": "1"
									},
									"targetFlag": {
										"name": "test-flag"
									},
									"value": true
								}
							]
						}
					}
				`,
			},
		})
	})

}

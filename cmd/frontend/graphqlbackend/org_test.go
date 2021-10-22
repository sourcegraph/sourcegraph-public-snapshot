package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrganization(t *testing.T) {
	resetMocks()
	database.Mocks.Orgs.GetByName = func(context.Context, string) (*types.Org, error) {
		return &types.Org{ID: 1, Name: "acme"}, nil
	}

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					organization(name: "acme") {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"organization": {
						"name": "acme"
					}
				}
			`,
		},
	})
}

func TestOrganizationRepositories(t *testing.T) {
	resetMocks()
	database.Mocks.Orgs.GetByName = func(context.Context, string) (*types.Org, error) {
		return &types.Org{ID: 1, Name: "acme"}, nil
	}
	database.Mocks.Repos.List = func(context.Context, database.ReposListOptions) (repos []*types.Repo, err error) {
		return []*types.Repo{
			{
				Name: "acme-repo",
			},
		}, nil
	}
	database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	database.Mocks.OrgMembers.GetByOrgIDAndUserID = func(ctx context.Context, orgID, userID int32) (*types.OrgMembership, error) {
		return &types.OrgMembership{
			OrgID:  1,
			UserID: 1,
		}, nil
	}
	database.Mocks.FeatureFlags.GetOrgFeatureFlag = func(ctx context.Context, orgID int32, flagName string) (bool, error) {
		return true, nil
	}

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	defer func() {
		resetMocks()
	}()

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					organization(name: "acme") {
						name,
						repositories {
							nodes {
								name
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"organization": {
						"name": "acme",
						"repositories": {
							"nodes": [{
								"name": "acme-repo"
							}]
						}
					}
				}
			`,
			Context: ctx,
		},
	})
}

func TestNode_Org(t *testing.T) {
	resetMocks()
	database.Mocks.Orgs.MockGetByID_Return(t, &types.Org{ID: 1, Name: "acme"}, nil)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					node(id: "T3JnOjE=") {
						id
						... on Org {
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "T3JnOjE=",
						"name": "acme"
					}
				}
			`,
		},
	})
}

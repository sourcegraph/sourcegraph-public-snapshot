package resolvers

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrganizationRepositories(t *testing.T) {
	orgs := dbmock.NewMockOrgStore()
	orgs.GetByNameFunc.SetDefaultReturn(&types.Org{ID: 1, Name: "acme"}, nil)
	repo := &types.Repo{
		Name: "acme-repo",
	}
	database.Mocks.Repos.List = func(context.Context, database.ReposListOptions) (repos []*types.Repo, err error) {
		return []*types.Repo{
			repo,
		}, nil
	}

	users := dbmock.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmock.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(&types.OrgMembership{OrgID: 1, UserID: 1}, nil)

	featureFlags := dbmock.NewMockFeatureFlagStore()
	featureFlags.GetOrgFeatureFlagFunc.SetDefaultReturn(true, nil)

	db := dbmock.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	graphqlbackend.RunTests(t, []*graphqlbackend.Test{
		{
			Schema: func() *graphql.Schema {
				t.Helper()

				parsedSchema, parseSchemaErr := graphqlbackend.NewSchema(db, nil, nil, nil, nil, nil, nil, nil, nil, &mockEnterpriseResolver{db: db, repo: repo})
				if parseSchemaErr != nil {
					t.Fatal(parseSchemaErr)
				}

				return parsedSchema
			}(),
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

type mockEnterpriseResolver struct {
	db   database.DB
	repo *types.Repo
}

func (r *mockEnterpriseResolver) OrgRepositories(ctx context.Context, args *graphqlbackend.ListOrgRepositoriesArgs, org *types.Org) (graphqlbackend.RepositoryConnectionResolver, error) {
	return r, nil
}

func (r mockEnterpriseResolver) Nodes(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	return []*graphqlbackend.RepositoryResolver{graphqlbackend.NewRepositoryResolver(r.db, r.repo)}, nil
}

func (r mockEnterpriseResolver) TotalCount(ctx context.Context, args *graphqlbackend.TotalCountArgs) (*int32, error) {
	one := int32(1)
	return &one, nil
}

func (r mockEnterpriseResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return nil, nil
}

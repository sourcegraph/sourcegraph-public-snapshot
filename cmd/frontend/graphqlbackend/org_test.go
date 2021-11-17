package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOrganization(t *testing.T) {
	users := dbmock.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmock.NewMockOrgMemberStore()
	orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, nil)

	orgs := dbmock.NewMockOrgStore()
	orgs.GetByNameFunc.SetDefaultReturn(&types.Org{ID: 1, Name: "acme"}, nil)

	db := dbmock.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)

	t.Run("anyone can access by default", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
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
	})

	t.Run("users not invited or not a member cannot access on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
				{
					organization(name: "acme") {
						name
					}
				}
			`,
				ExpectedResult: `
				{
					"organization": null
				}
				`,
				ExpectedErrors: []*errors.QueryError{
					{
						Message: "org not found: name acme",
						Path:    []interface{}{"organization"},
					},
				},
			},
		})
	})

	t.Run("org members can access on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		users := dbmock.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)

		orgMembers := dbmock.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(&types.OrgMembership{OrgID: 1, UserID: 1}, nil)

		db := dbmock.NewMockDBFrom(db)
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgMembersFunc.SetDefaultReturn(orgMembers)

		RunTests(t, []*Test{
			{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
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
	})

	t.Run("invited users can access on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

		users := dbmock.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: false}, nil)

		orgMembers := dbmock.NewMockOrgMemberStore()
		orgMembers.GetByOrgIDAndUserIDFunc.SetDefaultReturn(nil, &database.ErrOrgMemberNotFound{})

		orgInvites := dbmock.NewMockOrgInvitationStore()
		orgInvites.GetPendingFunc.SetDefaultReturn(nil, nil)

		db := dbmock.NewMockDBFrom(db)
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgMembersFunc.SetDefaultReturn(orgMembers)
		db.OrgInvitationsFunc.SetDefaultReturn(orgInvites)

		RunTests(t, []*Test{
			{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
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
	})
}

func TestOrganizationRepositories_OSS(t *testing.T) {
	db := dbmock.NewMockDB()
	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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
			ExpectedErrors: []*errors.QueryError{{
				Message:   `Cannot query field "repositories" on type "Org".`,
				Locations: []errors.Location{{Line: 5, Column: 7}},
				Rule:      "FieldsOnCorrectType",
			}},
			Context: ctx,
		},
	})
}

func TestNode_Org(t *testing.T) {
	orgs := dbmock.NewMockOrgStore()
	orgs.GetByIDFunc.SetDefaultReturn(&types.Org{ID: 1, Name: "acme"}, nil)

	db := dbmock.NewMockDB()
	db.OrgsFunc.SetDefaultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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

func TestUnmarshalOrgID(t *testing.T) {
	t.Run("Valid org ID is parsed correctly", func(t *testing.T) {
		const id = int32(1)
		namespaceOrgID := relay.MarshalID("Org", id)
		orgID, err := UnmarshalOrgID(namespaceOrgID)
		assert.NoError(t, err)
		assert.Equal(t, id, orgID)
	})

	t.Run("Returns error for invalid org ID", func(t *testing.T) {
		const id = 1
		namespaceOrgID := relay.MarshalID("User", id)
		_, err := UnmarshalOrgID(namespaceOrgID)
		assert.Error(t, err)
	})
}

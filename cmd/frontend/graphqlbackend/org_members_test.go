package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAutocompleteMembersSearch(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	autocompleteResults := []*types.OrgMemberAutocompleteSearchItem{
		{
			ID:    1,
			InOrg: 1,
		},
		{
			ID:    2,
			InOrg: 0,
		},
		{
			ID:    3,
			InOrg: 0,
		},
	}
	orgMembers.AutocompleteMembersSearchFunc.SetDefaultReturn(autocompleteResults, nil)

	db := dbmocks.NewMockDB()
	//db.OrgsFunc.SetDefaultReturn(orgs)
	db.UsersFunc.SetDefaultReturn(users)
	db.OrgMembersFunc.SetDefaultReturn(orgMembers)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})

	t.Run("Returns expected results in the response", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schema:  mustParseGraphQLSchema(t, db),
				Context: ctx,
				Query: `
				query AutocompleteMembersSearch($organization: ID!, $query: String!) {
					autocompleteMembersSearch(organization: $organization, query: $query) {
						id
						inOrg
					}
				}
				`,
				Variables: map[string]any{
					"organization": string(MarshalOrgID(1)),
					"query":        "test",
				},
				ExpectedResult: fmt.Sprintf(`{
					"autocompleteMembersSearch": [
						{ "id": "%s","inOrg": %t },
						{ "id": "%s","inOrg": %t },
						{ "id": "%s","inOrg": %t }
					]
				}`,
					string(MarshalUserID(autocompleteResults[0].ID)),
					true,
					string(MarshalUserID(autocompleteResults[1].ID)),
					false,
					string(MarshalUserID(autocompleteResults[2].ID)),
					false),
			},
		})
	})
}

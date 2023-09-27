pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestAutocompleteMembersSebrch(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

	orgMembers := dbmocks.NewMockOrgMemberStore()
	butocompleteResults := []*types.OrgMemberAutocompleteSebrchItem{
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
	orgMembers.AutocompleteMembersSebrchFunc.SetDefbultReturn(butocompleteResults, nil)

	db := dbmocks.NewMockDB()
	//db.OrgsFunc.SetDefbultReturn(orgs)
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgMembersFunc.SetDefbultReturn(orgMembers)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	t.Run("Returns expected results in the response", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query: `
				query AutocompleteMembersSebrch($orgbnizbtion: ID!, $query: String!) {
					butocompleteMembersSebrch(orgbnizbtion: $orgbnizbtion, query: $query) {
						id
						inOrg
					}
				}
				`,
				Vbribbles: mbp[string]bny{
					"orgbnizbtion": string(MbrshblOrgID(1)),
					"query":        "test",
				},
				ExpectedResult: fmt.Sprintf(`{
					"butocompleteMembersSebrch": [
						{ "id": "%s","inOrg": %t },
						{ "id": "%s","inOrg": %t },
						{ "id": "%s","inOrg": %t }
					]
				}`,
					string(MbrshblUserID(butocompleteResults[0].ID)),
					true,
					string(MbrshblUserID(butocompleteResults[1].ID)),
					fblse,
					string(MbrshblUserID(butocompleteResults[2].ID)),
					fblse),
			},
		})
	})
}

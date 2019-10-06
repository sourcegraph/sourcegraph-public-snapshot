package campaigns

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_CampaignPreview(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	mocks.campaigns.Create = func(campaign *dbCampaign) (*dbCampaign, error) {
		t.Fatal("want campaign to not be persisted in the DB")
		return nil, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				 query {
					campaignPreview(input: { campaign: { namespace: "T3JnOjE=", name: "n", extensionData: { rawDiagnostics: [], rawFileDiffs: [] } } }) {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"campaignPreview": {
						"name": "n"
					}
				}
			`,
		},
	})
}

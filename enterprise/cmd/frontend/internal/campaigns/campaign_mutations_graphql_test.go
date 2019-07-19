package campaigns

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
)

func TestGraphQL_CreateCampaign(t *testing.T) {
	resetMocks()
	const wantProjectID = 1
	projects.MockProjectByDBID = func(id int64) (graphqlbackend.Project, error) {
		return projects.TestNewProject(wantProjectID, "", 0, 0), nil
	}
	wantCampaign := &dbCampaign{
		ProjectID:   wantProjectID,
		Name:        "n",
		Description: strptr("d"),
	}
	mocks.campaigns.Create = func(campaign *dbCampaign) (*dbCampaign, error) {
		if !reflect.DeepEqual(campaign, wantCampaign) {
			t.Errorf("got campaign %+v, want %+v", campaign, wantCampaign)
		}
		tmp := *campaign
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					campaigns {
						createCampaign(input: { project: "T3JnOjE=", name: "n", description: "d" }) {
							id
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"campaigns": {
						"createCampaign": {
							"id": "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==",
							"name": "n"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateCampaign(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.campaigns.GetByID = func(id int64) (*dbCampaign, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbCampaign{ID: wantID}, nil
	}
	mocks.campaigns.Update = func(id int64, update dbCampaignUpdate) (*dbCampaign, error) {
		if want := (dbCampaignUpdate{Name: strptr("n1"), Description: strptr("d1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbCampaign{
			ID:          2,
			ProjectID:   1,
			Name:        "n1",
			Description: strptr("d1"),
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					campaigns {
						updateCampaign(input: { id: "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==", name: "n1", description: "d1" }) {
							id
							name
							description
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"campaigns": {
						"updateCampaign": {
							"id": "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==",
							"name": "n1",
							"description": "d1"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteCampaign(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.campaigns.GetByID = func(id int64) (*dbCampaign, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbCampaign{ID: wantID}, nil
	}
	mocks.campaigns.DeleteByID = func(id int64) error {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					campaigns {
						deleteCampaign(campaign: "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==") {
							alwaysNil
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"campaigns": {
						"deleteCampaign": null
					}
				}
			`,
		},
	})
}

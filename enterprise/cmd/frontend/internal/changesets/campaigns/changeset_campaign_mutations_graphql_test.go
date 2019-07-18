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

func TestGraphQL_CreateChangesetCampaign(t *testing.T) {
	resetMocks()
	const wantProjectID = 1
	projects.MockProjectByDBID = func(id int64) (graphqlbackend.Project, error) {
		return projects.TestNewProject(wantProjectID, "", 0, 0), nil
	}
	wantChangesetCampaign := &dbChangesetCampaign{
		ProjectID:   wantProjectID,
		Name:        "n",
		Description: strptr("d"),
	}
	mocks.campaigns.Create = func(changesetCampaign *dbChangesetCampaign) (*dbChangesetCampaign, error) {
		if !reflect.DeepEqual(changesetCampaign, wantChangesetCampaign) {
			t.Errorf("got changesetCampaign %+v, want %+v", changesetCampaign, wantChangesetCampaign)
		}
		tmp := *changesetCampaign
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					changesetCampaigns {
						createChangesetCampaign(input: { project: "T3JnOjE=", name: "n", description: "d" }) {
							id
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesetCampaigns": {
						"createChangesetCampaign": {
							"id": "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==",
							"name": "n"
						}
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateChangesetCampaign(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.campaigns.GetByID = func(id int64) (*dbChangesetCampaign, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbChangesetCampaign{ID: wantID}, nil
	}
	mocks.campaigns.Update = func(id int64, update dbChangesetCampaignUpdate) (*dbChangesetCampaign, error) {
		if want := (dbChangesetCampaignUpdate{Name: strptr("n1"), Description: strptr("d1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbChangesetCampaign{
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
					changesetCampaigns {
						updateChangesetCampaign(input: { id: "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==", name: "n1", description: "d1" }) {
							id
							name
							description
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesetCampaigns": {
						"updateChangesetCampaign": {
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

func TestGraphQL_DeleteChangesetCampaign(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.campaigns.GetByID = func(id int64) (*dbChangesetCampaign, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbChangesetCampaign{ID: wantID}, nil
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
					changesetCampaigns {
						deleteChangesetCampaign(changesetCampaign: "Q2hhbmdlc2V0Q2FtcGFpZ246Mg==") {
							alwaysNil
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"changesetCampaigns": {
						"deleteChangesetCampaign": null
					}
				}
			`,
		},
	})
}

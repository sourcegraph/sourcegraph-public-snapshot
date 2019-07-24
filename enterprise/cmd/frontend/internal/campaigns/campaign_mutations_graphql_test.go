package campaigns

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"

	_ "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func TestGraphQL_CreateCampaign(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	wantCampaign := &dbCampaign{
		NamespaceOrgID: wantOrgID,
		Name:           "n",
		Description:    strptr("d"),
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
						createCampaign(input: { namespace: "T3JnOjE=", name: "n", description: "d" }) {
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
							"id": "Q2FtcGFpZ246Mg==",
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
			ID:             2,
			NamespaceOrgID: 1,
			Name:           "n1",
			Description:    strptr("d1"),
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					campaigns {
						updateCampaign(input: { id: "Q2FtcGFpZ246Mg==", name: "n1", description: "d1" }) {
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
							"id": "Q2FtcGFpZ246Mg==",
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
						deleteCampaign(campaign: "Q2FtcGFpZ246Mg==") {
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

func TestGraphQL_AddRemoveThreadsToFromCampaign(t *testing.T) {
	resetMocks()
	const wantCampaignID = 2
	mocks.campaigns.GetByID = func(id int64) (*dbCampaign, error) {
		if id != wantCampaignID {
			t.Errorf("got ID %d, want %d", id, wantCampaignID)
		}
		return &dbCampaign{ID: wantCampaignID}, nil
	}
	const wantThreadID = 3
	db.Mocks.DiscussionThreads.Get = func(int64) (*types.DiscussionThread, error) {
		return &types.DiscussionThread{ID: wantThreadID}, nil
	}
	addRemoveThreadsToFromCampaign := func(campaign int64, threads []int64) error {
		if campaign != wantCampaignID {
			t.Errorf("got %d, want %d", campaign, wantCampaignID)
		}
		if want := []int64{wantThreadID}; !reflect.DeepEqual(threads, want) {
			t.Errorf("got %v, want %v", threads, want)
		}
		return nil
	}

	tests := map[string]*func(thread int64, threads []int64) error{
		"addThreadsToCampaign":      &mocks.campaignsThreads.AddThreadsToCampaign,
		"removeThreadsFromCampaign": &mocks.campaignsThreads.RemoveThreadsFromCampaign,
	}
	for name, mockFn := range tests {
		t.Run(name, func(t *testing.T) {
			*mockFn = addRemoveThreadsToFromCampaign
			defer func() { *mockFn = nil }()

			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Context: backend.WithAuthzBypass(context.Background()),
					Schema:  graphqlbackend.GraphQLSchema,
					Query: `
				mutation {
					campaigns {
						` + name + `(campaign: "Q2FtcGFpZ246Mg==", threads: ["RGlzY3Vzc2lvblRocmVhZDoiMyI="]) {
							__typename
						}
					}
				}
			`,
					ExpectedResult: `
				{
					"campaigns": {
						"` + name + `": {
							"__typename": "EmptyResponse"
						}
					}
				}
			`,
				},
			})
		})
	}
}

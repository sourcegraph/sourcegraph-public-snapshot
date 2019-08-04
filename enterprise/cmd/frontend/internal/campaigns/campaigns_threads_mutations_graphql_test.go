package campaigns

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestGraphQL_AddRemoveThreadsToFromCampaign(t *testing.T) {
	resetMocks()
	const wantCampaignID = 2
	mocks.campaigns.GetByID = func(id int64) (*dbCampaign, error) {
		if id != wantCampaignID {
			t.Errorf("got ID %d, want %d", id, wantCampaignID)
		}
		return &dbCampaign{ID: wantCampaignID}, nil
	}
	events.MockCreateEvent = func(event events.CreationData) error {
		if event.Objects.Campaign != wantCampaignID {
			t.Errorf("got campaign ID %d, want %d", event.Objects.Campaign, wantCampaignID)
		}
		return nil
	}
	defer func() { events.MockCreateEvent = nil }()
	const wantThreadID = 3
	mockGetThreadDBIDs = func([]graphql.ID) ([]int64, error) { return []int64{wantThreadID}, nil }
	defer func() { mockGetThreadDBIDs = nil }()
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
					` + name + `(campaign: "Q2FtcGFpZ246Mg==", threads: ["RGlzY3Vzc2lvblRocmVhZDoiMyI="]) {
						__typename
					}
				}
			`,
					ExpectedResult: `
				{
					"` + name + `": {
						"__typename": "EmptyResponse"
					}
				}
			`,
				},
			})
		})
	}
}

package events

import (
	"context"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestGraphQL_CreateEvent(t *testing.T) {
	resetMocks()
	const wantOrgID = 1
	db.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1}, nil
	}
	db.Mocks.Orgs.GetByID = func(context.Context, int32) (*types.Org, error) {
		return &types.Org{ID: wantOrgID}, nil
	}
	wantEvent := &dbEvent{
		NamespaceOrgID: wantOrgID,
		Name:           "n",
	}
	mocks.events.Create = func(event *dbEvent) (*dbEvent, error) {
		if !reflect.DeepEqual(event, wantEvent) {
			t.Errorf("got event %+v, want %+v", event, wantEvent)
		}
		tmp := *event
		tmp.ID = 2
		return &tmp, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					createEvent(input: { namespace: "T3JnOjE=", name: "n" }) {
						id
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"createEvent": {
						"id": "Q2FtcGFpZ246Mg==",
						"name": "n"
					}
				}
			`,
		},
	})
}

func TestGraphQL_UpdateEvent(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.events.GetByID = func(id int64) (*dbEvent, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbEvent{ID: wantID}, nil
	}
	mocks.events.Update = func(id int64, update dbEventUpdate) (*dbEvent, error) {
		if want := (dbEventUpdate{Name: strptr("n1")}); !reflect.DeepEqual(update, want) {
			t.Errorf("got update %+v, want %+v", update, want)
		}
		return &dbEvent{
			ID:             2,
			NamespaceOrgID: 1,
			Name:           "n1",
		}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					updateEvent(input: { id: "Q2FtcGFpZ246Mg==", name: "n1" }) {
						id
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"updateEvent": {
						"id": "Q2FtcGFpZ246Mg==",
						"name": "n1"
					}
				}
			`,
		},
	})
}

func TestGraphQL_DeleteEvent(t *testing.T) {
	resetMocks()
	const wantID = 2
	mocks.events.GetByID = func(id int64) (*dbEvent, error) {
		if id != wantID {
			t.Errorf("got ID %d, want %d", id, wantID)
		}
		return &dbEvent{ID: wantID}, nil
	}
	mocks.events.DeleteByID = func(id int64) error {
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
					deleteEvent(event: "Q2FtcGFpZ246Mg==") {
						alwaysNil
					}
				}
			`,
			ExpectedResult: `
				{
					"deleteEvent": null
				}
			`,
		},
	})
}

func TestGraphQL_AddRemoveThreadsToFromEvent(t *testing.T) {
	resetMocks()
	const wantEventID = 2
	mocks.events.GetByID = func(id int64) (*dbEvent, error) {
		if id != wantEventID {
			t.Errorf("got ID %d, want %d", id, wantEventID)
		}
		return &dbEvent{ID: wantEventID}, nil
	}
	const wantThreadID = 3
	mockGetThreadDBIDs = func([]graphql.ID) ([]int64, error) { return []int64{wantThreadID}, nil }
	defer func() { mockGetThreadDBIDs = nil }()
	addRemoveThreadsToFromEvent := func(event int64, threads []int64) error {
		if event != wantEventID {
			t.Errorf("got %d, want %d", event, wantEventID)
		}
		if want := []int64{wantThreadID}; !reflect.DeepEqual(threads, want) {
			t.Errorf("got %v, want %v", threads, want)
		}
		return nil
	}

	tests := map[string]*func(thread int64, threads []int64) error{
		"addThreadsToEvent":      &mocks.eventsThreads.AddThreadsToEvent,
		"removeThreadsFromEvent": &mocks.eventsThreads.RemoveThreadsFromEvent,
	}
	for name, mockFn := range tests {
		t.Run(name, func(t *testing.T) {
			*mockFn = addRemoveThreadsToFromEvent
			defer func() { *mockFn = nil }()

			gqltesting.RunTests(t, []*gqltesting.Test{
				{
					Context: backend.WithAuthzBypass(context.Background()),
					Schema:  graphqlbackend.GraphQLSchema,
					Query: `
				mutation {
					` + name + `(event: "Q2FtcGFpZ246Mg==", threads: ["RGlzY3Vzc2lvblRocmVhZDoiMyI="]) {
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

package diagnostics

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

type mockThread struct {
	graphqlbackend.Thread
	id int64
}

func (t mockThread) ID() graphql.ID { return graphqlbackend.MarshalThreadID(t.id) }

func TestGraphQL_AddDiagnosticsToThread(t *testing.T) {
	resetMocks()
	const (
		wantThreadDiagnosticID = 2
		wantThreadID           = 3
	)
	threads.MockThreadByID = func(id graphql.ID) (graphqlbackend.Thread, error) {
		return mockThread{id: wantThreadID}, nil
	}
	defer func() { threads.MockThreadByID = nil }()
	mocks.threadsDiagnostics.Create = func(threadDiagnostic dbThreadDiagnostic) (int64, error) {
		return wantThreadDiagnosticID, nil
	}
	events.MockCreateEvent = func(event events.CreationData) error {
		if event.Objects.Thread != wantThreadID {
			t.Errorf("got thread ID %d, want %d", event.Objects.Thread, wantThreadID)
		}
		if event.Objects.ThreadDiagnosticEdge != wantThreadDiagnosticID {
			t.Errorf("got thread diagnostic ID %d, want %d", event.Objects.ThreadDiagnosticEdge, wantThreadDiagnosticID)
		}
		return nil
	}
	defer func() { events.MockCreateEvent = nil }()

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation {
					addDiagnosticsToThread(thread: "RGlzY3Vzc2lvblRocmVhZDoiMyI=", rawDiagnostics: ["{}", "{}"]) {
						__typename
					}
				}
			`,
			ExpectedResult: `
				{
					"addDiagnosticsToThread": {
						"__typename": "ThreadDiagnosticConnection"
					}
				}
			`,
		},
	})
}

func TestGraphQL_RemoveDiagnosticsFromThread(t *testing.T) {
	resetMocks()
	const (
		wantThreadDiagnosticID = 2
		wantThreadID           = 3
	)
	threads.MockThreadByID = func(id graphql.ID) (graphqlbackend.Thread, error) {
		return mockThread{id: wantThreadID}, nil
	}
	defer func() { threads.MockThreadByID = nil }()
	mocks.threadsDiagnostics.DeleteByID = func(threadDiagnosticID, threadID int64) error {
		if threadDiagnosticID != wantThreadDiagnosticID {
			t.Errorf("got %d, want %d", threadDiagnosticID, wantThreadDiagnosticID)
		}
		if threadID != wantThreadID {
			t.Errorf("got %d, want %d", threadID, wantThreadID)
		}
		return nil
	}
	mocks.threadsDiagnostics.GetByID = func(threadDiagnosticID int64) (*dbThreadDiagnostic, error) {
		if threadDiagnosticID != wantThreadDiagnosticID {
			t.Errorf("got %d, want %d", threadDiagnosticID, wantThreadDiagnosticID)
		}
		return &dbThreadDiagnostic{ID: wantThreadDiagnosticID, ThreadID: wantThreadID}, nil
	}
	events.MockCreateEvent = func(event events.CreationData) error {
		if event.Objects.Thread != wantThreadID {
			t.Errorf("got thread ID %d, want %d", event.Objects.Thread, wantThreadID)
		}
		if event.Objects.ThreadDiagnosticEdge != 0 {
			t.Errorf("got thread diagnostic ID %d, want 0", event.Objects.ThreadDiagnosticEdge)
		}
		return nil
	}
	defer func() { events.MockCreateEvent = nil }()

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: backend.WithAuthzBypass(context.Background()),
			Schema:  graphqlbackend.GraphQLSchema,
			Query: `
				mutation($thread: ID!, $threadDiagnostic: ID!) {
					removeDiagnosticsFromThread(thread: $thread, threadDiagnosticEdges: [$threadDiagnostic]) {
						__typename
					}
				}
			`,
			Variables: map[string]interface{}{
				"thread":           string(graphqlbackend.MarshalThreadID(wantThreadID)),
				"threadDiagnostic": string(graphqlbackend.MarshalThreadDiagnosticEdgeID(wantThreadDiagnosticID)),
			},
			ExpectedResult: `
				{
					"removeDiagnosticsFromThread": {
						"__typename": "EmptyResponse"
					}
				}
			`,
		},
	})
}

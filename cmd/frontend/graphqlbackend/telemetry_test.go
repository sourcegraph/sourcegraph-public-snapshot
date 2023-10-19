package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type mockTelemetryResolver struct {
	events []TelemetryEventInput
}

func (m *mockTelemetryResolver) RecordEvents(_ context.Context, args *RecordEventsArgs) (*EmptyResponse, error) {
	m.events = append(m.events, args.Events...)
	return &EmptyResponse{}, nil
}

func TestTelemetryRecordEvents(t *testing.T) {
	mockResolver := &mockTelemetryResolver{}
	parsedSchema, err := NewSchema(
		dbmocks.NewMockDB(),
		gitserver.NewClient(),
		[]OptionalResolver{{
			TelemetryRootResolver: &TelemetryRootResolver{Resolver: mockResolver},
		}},
		graphql.PanicHandler(printStackTrace{&gqlerrors.DefaultPanicHandler{}}),
	)
	require.NoError(t, err)

	// Write a raw GraphQL event because we want to test providing the raw input
	// value, as if from a client, which the Variables field in RunTest doesn't
	// seem to accept right (it wants the final type, which defeats the point)
	gqlEventInput := `
	{
		feature: "cody.fixup"
		action: "applied"
		source: {
		  client: "VSCode.Cody",
		  clientVersion: "0.14.1"
		}
		parameters: {
		  version: 0
		  metadata: [
			{
			  key: "contextSelection",
			  value: 1
			},
			{
			  key: "chatPredictions",
			  value: 0
			},
		  ]
		  privateMetadata: {key:"value"}
		}
	  }
	`

	// Check all fields accepted in GraphQL resolver.
	RunTest(t, &Test{
		Schema:  parsedSchema,
		Context: context.Background(),
		Query: fmt.Sprintf(`mutation RecordTelemetryEvents() {
			telemetry {
				recordEvents(events: [%s]) {
					alwaysNil
				}
			}
		}`, gqlEventInput),
		ExpectedResult: `{
			"telemetry": {
				"recordEvents": {
					"alwaysNil": null
				}
			}
		}`,
	})

}

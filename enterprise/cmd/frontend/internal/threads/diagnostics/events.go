package diagnostics

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
)

const (
	eventTypeAddDiagnosticToThread      events.Type = "AddDiagnosticToThread"
	eventTypeRemoveDiagnosticFromThread             = "RemoveDiagnosticFromThread"
)

func init() {
	for _, eventType_ := range []events.Type{eventTypeAddDiagnosticToThread, eventTypeRemoveDiagnosticFromThread} {
		eventType := eventType_
		events.Register(eventType, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
			thread, err := graphqlbackend.ThreadByID(ctx, graphqlbackend.MarshalThreadID(data.Thread))
			if err != nil {
				return err
			}
			event := &graphqlbackend.AddRemoveDiagnosticToFromThreadEvent{
				EventCommon: common,
				Thread_:     thread,
			}

			// Store the diagnostic in the event data so that it is still accessible even if the
			// diagnostic is removed from the thread.
			var o diagnostics.GQLDiagnostic
			if err := json.Unmarshal(data.Data, &o); err != nil {
				return err
			}
			event.Diagnostic_ = o

			switch {
			case eventType == eventTypeAddDiagnosticToThread:
				edge, err := dbThreadDiagnosticEdges{}.GetByID(ctx, data.ThreadDiagnosticEdge)
				if err != nil && err != errThreadDiagnosticNotFound {
					return err
				}
				if edge != nil {
					event.Edge_ = &gqlThreadDiagnosticEdge{edge}
				}
				toEvent.AddDiagnosticToThreadEvent = event
			case eventType == eventTypeRemoveDiagnosticFromThread:
				toEvent.RemoveDiagnosticFromThreadEvent = event
			default:
				panic("unexpected event type: " + eventType)
			}
			return nil
		})
	}
}

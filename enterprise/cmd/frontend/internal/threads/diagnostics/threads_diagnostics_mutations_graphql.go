package diagnostics

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) AddDiagnosticsToThread(ctx context.Context, arg *graphqlbackend.AddDiagnosticsToThreadArgs) (graphqlbackend.ThreadDiagnosticConnection, error) {
	// 🚨 SECURITY: Any viewer can add/remove diagnostics to/from a thread. TODO!(sqs)
	thread, err := threads.GraphQLResolver{}.ThreadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	threadID, err := graphqlbackend.UnmarshalThreadID(thread.ID())
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(arg.RawDiagnostics))
	for i, rawDiagnostic := range arg.RawDiagnostics {
		var d diagnostics.GQLDiagnostic
		if err := json.Unmarshal([]byte(rawDiagnostic), &d); err != nil {
			return nil, err
		}
		id, err := (dbThreadDiagnosticEdges{}).Create(ctx, dbThreadDiagnostic{
			ThreadID: threadID,
			Type:     d.Type_,
			Data:     d.Data_,
		})
		if err != nil {
			return nil, err
		}
		ids[i] = id

		if err := events.CreateEvent(ctx, nil, events.CreationData{
			Type: eventTypeAddDiagnosticToThread,
			Objects: events.Objects{
				Thread:               threadID,
				ThreadDiagnosticEdge: id,
			},
			Data: d,
		}); err != nil {
			return nil, err
		}
	}
	return &threadDiagnosticConnection{
		opt: dbThreadDiagnosticEdgesListOptions{IDs: ids},
	}, nil
}

func (GraphQLResolver) RemoveDiagnosticsFromThread(ctx context.Context, arg *graphqlbackend.RemoveDiagnosticsFromThreadArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Any viewer can add/remove diagnostics to/from a thread. TODO!(sqs)
	// 🚨 SECURITY: Any viewer can add/remove diagnostics to/from a thread. TODO!(sqs)
	thread, err := threads.GraphQLResolver{}.ThreadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	threadID, err := graphqlbackend.UnmarshalThreadID(thread.ID())
	if err != nil {
		return nil, err
	}

	for _, threadDiagnosticGQLID := range arg.ThreadDiagnosticEdges {
		threadDiagnosticID, err := graphqlbackend.UnmarshalThreadDiagnosticEdgeID(threadDiagnosticGQLID)
		if err != nil {
			return nil, err
		}

		// Get the edge so we can create an event for this removal with its data (before deleting
		// it).
		edge, err := (dbThreadDiagnosticEdges{}).GetByID(ctx, threadDiagnosticID)
		if err != nil {
			return nil, err
		}

		if err := (dbThreadDiagnosticEdges{}).DeleteByIDInThread(ctx, threadDiagnosticID, threadID); err != nil {
			return nil, err
		}

		if err := events.CreateEvent(ctx, nil, events.CreationData{
			Type: eventTypeRemoveDiagnosticFromThread,
			Objects: events.Objects{
				Thread: threadID,
			},
			Data: diagnostics.GQLDiagnostic{Type_: edge.Type, Data_: edge.Data},
		}); err != nil {
			return nil, err
		}

	}

	return &graphqlbackend.EmptyResponse{}, nil
}

package diagnostics

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
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

	// TODO!(sqs): record events
	//
	// if err := events.CreateEvent(ctx, events.CreationData{
	// 	Type:    eventType,
	// 	Objects: events.Objects{Thread: threadDBID},
	// }); err != nil {
	// 	return nil, err
	// }

	ids := make([]int64, len(arg.RawDiagnostics))
	for i, rawDiagnostic := range arg.RawDiagnostics {
		const dummytype = "TYPE TODO!(sqs)"
		id, err := (dbThreadsDiagnostics{}).Create(ctx, dbThreadDiagnostic{
			ThreadID: threadID,
			Type:     dummytype,
			Data:     json.RawMessage(rawDiagnostic),
		})
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return &threadDiagnosticConnection{
		opt: dbThreadsDiagnosticsListOptions{IDs: ids},
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
		if err := (dbThreadsDiagnostics{}).DeleteByIDInThread(ctx, threadDiagnosticID, threadID); err != nil {
			return nil, err
		}
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

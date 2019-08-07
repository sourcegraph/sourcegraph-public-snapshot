package diagnostics

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) Create(ctx context.Context, arg *graphqlbackend.CreateArgs) (*graphqlbackend.EmptyResponse, error) {
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

	for _, rawDiagnostic := range arg.RawDiagnostics {
		const dummytype = "TYPE TODO!(sqs)"
		dbDiagnostic := &dbThreadDiagnostic{
			ThreadID: threadID,
			Type:     dummytype,
			Data:     json.RawMessage(rawDiagnostic),
		}
		if err := (dbThreadDiagnostics{}).Create(ctx, dbDiagnostic); err != nil {
			return err
		}
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (GraphQLResolver) Delete(ctx context.Context, arg *graphqlbackend.DeleteArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Any viewer can add/remove diagnostics to/from a thread. TODO!(sqs)
	if err := addRemoveDiagnosticsToFromThread(ctx, arg.Campaign, nil, arg.Threads); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func addRemoveDiagnosticsToFromThread(ctx context.Context, campaignID graphql.ID, addThreads []graphql.ID, removeThreads []graphql.ID) error {

	campaign, err := campaignByID(ctx, campaignID)
	if err != nil {
		return err
	}

	if len(addThreads) > 0 {
		addThreadDBIDs, err := getThreadDBIDs(ctx, addThreads)
		if err != nil {
			return err
		}

		for _, threadDBID := range addThreadDBIDs {
			if err := createEvent(ctx, eventTypeAddThreadToCampaign, threadDBID); err != nil {
				return err
			}
		}
	}

	if len(removeThreads) > 0 {
		removeThreadDBIDs, err := getThreadDBIDs(ctx, removeThreads)
		if err != nil {
			return err
		}
		if err := (dbThreadDiagnostics{}).Delete(ctx, campaign.db.ID, removeThreadDBIDs); err != nil {
			return err
		}
		for _, threadDBID := range removeThreadDBIDs {
			if err := createEvent(ctx, eventTypeRemoveThreadFromCampaign, threadDBID); err != nil {
				return err
			}
		}
	}

	return nil
}

var mockGetThreadDBIDs func(threadIDs []graphql.ID) ([]int64, error)

func getThreadDBIDs(ctx context.Context, threadIDs []graphql.ID) ([]int64, error) {
	if mockGetThreadDBIDs != nil {
		return mockGetThreadDBIDs(threadIDs)
	}

	dbIDs := make([]int64, len(threadIDs))
	for i, threadID := range threadIDs {
		// 🚨 SECURITY: Only organization members and site admins may create threads in an
		// organization. The threadByID function performs this check.
		thread, err := threads.GraphQLResolver{}.ThreadByID(ctx, threadID)
		if err != nil {
			return nil, err
		}
		dbIDs[i] = thread.DBID()
	}
	return dbIDs, nil
}

package campaigns

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) AddThreadsToCampaign(ctx context.Context, arg *graphqlbackend.AddRemoveThreadsToFromCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := addRemoveThreadsToFromCampaign(ctx, arg.Campaign, arg.Threads, nil); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (GraphQLResolver) RemoveThreadsFromCampaign(ctx context.Context, arg *graphqlbackend.AddRemoveThreadsToFromCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := addRemoveThreadsToFromCampaign(ctx, arg.Campaign, nil, arg.Threads); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func addRemoveThreadsToFromCampaign(ctx context.Context, campaignID graphql.ID, addThreads []graphql.ID, removeThreads []graphql.ID) error {
	// 🚨 SECURITY: Any viewer can add/remove threads to/from a campaign.
	campaign, err := campaignByID(ctx, campaignID)
	if err != nil {
		return err
	}

	createEvent := func(ctx context.Context, eventType events.Type, threadDBID int64) error {
		return events.CreateEvent(ctx, nil, events.CreationData{
			Type: eventType,
			Objects: events.Objects{
				Campaign: campaign.db.ID,
				Thread:   threadDBID,
			},
		})
	}

	if len(addThreads) > 0 {
		addThreadDBIDs, err := getThreadDBIDs(ctx, addThreads)
		if err != nil {
			return err
		}
		if err := (dbCampaignsThreads{}).AddThreadsToCampaign(ctx, campaign.db.ID, addThreadDBIDs); err != nil {
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
		if err := (dbCampaignsThreads{}).RemoveThreadsFromCampaign(ctx, campaign.db.ID, removeThreadDBIDs); err != nil {
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

package campaigns

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/threads"
)

func (GraphQLResolver) CreateCampaign(ctx context.Context, arg *graphqlbackend.CreateCampaignArgs) (graphqlbackend.Campaign, error) {
	v := &dbCampaign{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
	}

	var err error
	v.NamespaceUserID, v.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, arg.Input.Namespace)
	if err != nil {
		return nil, err
	}

	campaign, err := dbCampaigns{}.Create(ctx, v)
	if err != nil {
		return nil, err
	}
	return &gqlCampaign{db: campaign}, nil
}

func (GraphQLResolver) UpdateCampaign(ctx context.Context, arg *graphqlbackend.UpdateCampaignArgs) (graphqlbackend.Campaign, error) {
	l, err := campaignByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	campaign, err := dbCampaigns{}.Update(ctx, l.db.ID, dbCampaignUpdate{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
	})
	if err != nil {
		return nil, err
	}
	return &gqlCampaign{db: campaign}, nil
}

func (GraphQLResolver) PublishPreviewCampaign(ctx context.Context, arg *graphqlbackend.PublishPreviewCampaignArgs) (graphqlbackend.Campaign, error) {
	gqlCampaign, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}
	panic("TODO!(sqs)")
	return gqlCampaign, nil
}

func (GraphQLResolver) DeleteCampaign(ctx context.Context, arg *graphqlbackend.DeleteCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlCampaign, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}
	return nil, dbCampaigns{}.DeleteByID(ctx, gqlCampaign.db.ID)
}

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

	if len(addThreads) > 0 {
		addThreadIDs, err := getThreadDBIDs(ctx, addThreads)
		if err != nil {
			return err
		}
		if err := (dbCampaignsThreads{}).AddThreadsToCampaign(ctx, campaign.db.ID, addThreadIDs); err != nil {
			return err
		}
	}

	if len(removeThreads) > 0 {
		removeThreadIDs, err := getThreadDBIDs(ctx, removeThreads)
		if err != nil {
			return err
		}
		if err := (dbCampaignsThreads{}).RemoveThreadsFromCampaign(ctx, campaign.db.ID, removeThreadIDs); err != nil {
			return err
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

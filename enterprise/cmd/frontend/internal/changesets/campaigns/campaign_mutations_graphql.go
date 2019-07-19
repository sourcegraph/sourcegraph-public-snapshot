package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateCampaign(ctx context.Context, arg *graphqlbackend.CreateCampaignArgs) (graphqlbackend.Campaign, error) {
	project, err := graphqlbackend.ProjectByID(ctx, arg.Input.Project)
	if err != nil {
		return nil, err
	}

	campaign, err := dbCampaigns{}.Create(ctx, &dbCampaign{
		ProjectID:   project.DBID(),
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
	})
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

func (GraphQLResolver) DeleteCampaign(ctx context.Context, arg *graphqlbackend.DeleteCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlCampaign, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}
	return nil, dbCampaigns{}.DeleteByID(ctx, gqlCampaign.db.ID)
}

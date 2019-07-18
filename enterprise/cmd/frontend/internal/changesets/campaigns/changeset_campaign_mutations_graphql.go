package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateChangesetCampaign(ctx context.Context, arg *graphqlbackend.CreateChangesetCampaignArgs) (graphqlbackend.ChangesetCampaign, error) {
	project, err := graphqlbackend.ProjectByID(ctx, arg.Input.Project)
	if err != nil {
		return nil, err
	}

	changesetCampaign, err := dbChangesetCampaigns{}.Create(ctx, &dbChangesetCampaign{
		ProjectID:   project.DBID(),
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
	})
	if err != nil {
		return nil, err
	}
	return &gqlChangesetCampaign{db: changesetCampaign}, nil
}

func (GraphQLResolver) UpdateChangesetCampaign(ctx context.Context, arg *graphqlbackend.UpdateChangesetCampaignArgs) (graphqlbackend.ChangesetCampaign, error) {
	l, err := changesetCampaignByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	changesetCampaign, err := dbChangesetCampaigns{}.Update(ctx, l.db.ID, dbChangesetCampaignUpdate{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
	})
	if err != nil {
		return nil, err
	}
	return &gqlChangesetCampaign{db: changesetCampaign}, nil
}

func (GraphQLResolver) DeleteChangesetCampaign(ctx context.Context, arg *graphqlbackend.DeleteChangesetCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlChangesetCampaign, err := changesetCampaignByID(ctx, arg.ChangesetCampaign)
	if err != nil {
		return nil, err
	}
	return nil, dbChangesetCampaigns{}.DeleteByID(ctx, gqlChangesetCampaign.db.ID)
}

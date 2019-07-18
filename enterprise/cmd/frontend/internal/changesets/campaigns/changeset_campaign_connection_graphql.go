package campaigns

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (GraphQLResolver) ChangesetCampaignsDefinedIn(ctx context.Context, projectID graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ChangesetCampaignConnection, error) {
	// Check existence.
	project, err := graphqlbackend.ProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	list, err := dbChangesetCampaigns{}.List(ctx, dbChangesetCampaignsListOptions{ProjectID: project.DBID()})
	if err != nil {
		return nil, err
	}
	campaigns := make([]*gqlChangesetCampaign, len(list))
	for i, a := range list {
		campaigns[i] = &gqlChangesetCampaign{db: a}
	}
	return &changesetCampaignConnection{arg: arg, campaigns: campaigns}, nil
}

type changesetCampaignConnection struct {
	arg   *graphqlutil.ConnectionArgs
	campaigns []*gqlChangesetCampaign
}

func (r *changesetCampaignConnection) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetCampaign, error) {
	campaigns := r.campaigns
	if first := r.arg.First; first != nil && len(campaigns) > int(*first) {
		campaigns = campaigns[:int(*first)]
	}

	campaigns2 := make([]graphqlbackend.ChangesetCampaign, len(campaigns))
	for i, l := range campaigns {
		campaigns2[i] = l
	}
	return campaigns2, nil
}

func (r *changesetCampaignConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.campaigns)), nil
}

func (r *changesetCampaignConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.campaigns)), nil
}

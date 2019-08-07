package campaigns

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) Campaigns(ctx context.Context, arg *graphqlbackend.CampaignsArgs) (graphqlbackend.CampaignConnection, error) {
	var opt dbCampaignsListOptions
	if arg.Object != nil {
		threadID, err := threads.UnmarshalID(*arg.Object)
		if err != nil {
			return nil, err
		}
		opt.ObjectThreadID = threadID
	}
	return campaignsByOptions(ctx, opt, &arg.ConnectionArgs)
}

func (GraphQLResolver) CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CampaignConnection, error) {
	var opt dbCampaignsListOptions
	var err error
	opt.NamespaceUserID, opt.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return campaignsByOptions(ctx, opt, arg)
}

func (GraphQLResolver) CampaignsWithObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CampaignConnection, error) {
	return GraphQLResolver{}.Campaigns(ctx, &graphqlbackend.CampaignsArgs{Object: &object, ConnectionArgs: *arg})
}

func campaignsByOptions(ctx context.Context, opt dbCampaignsListOptions, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CampaignConnection, error) {
	list, err := dbCampaigns{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	campaigns := make([]*gqlCampaign, len(list))
	for i, a := range list {
		campaigns[i] = newGQLCampaign(a)
	}
	return &campaignConnection{arg: arg, campaigns: campaigns}, nil
}

type campaignConnection struct {
	arg       *graphqlutil.ConnectionArgs
	campaigns []*gqlCampaign
}

func (r *campaignConnection) Nodes(ctx context.Context) ([]graphqlbackend.Campaign, error) {
	campaigns := r.campaigns
	if first := r.arg.First; first != nil && len(campaigns) > int(*first) {
		campaigns = campaigns[:int(*first)]
	}

	campaigns2 := make([]graphqlbackend.Campaign, len(campaigns))
	for i, l := range campaigns {
		campaigns2[i] = l
	}
	return campaigns2, nil
}

func (r *campaignConnection) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.campaigns)), nil
}

func (r *campaignConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.arg.First != nil && int(*r.arg.First) < len(r.campaigns)), nil
}

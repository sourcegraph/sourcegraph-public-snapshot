package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rules"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) CreateCampaign(ctx context.Context, arg *graphqlbackend.CreateCampaignArgs) (graphqlbackend.Campaign, error) {
	v := &dbCampaign{
		Name: arg.Input.Name,
		// TODO!(sqs): description, renamed to body but allow it to be updated here
	}

	var err error
	v.NamespaceUserID, v.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, arg.Input.Namespace)
	if err != nil {
		return nil, err
	}

	authorUserID, err := comments.CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	comment := commentobjectdb.DBObjectCommentFields{Author: actor.DBColumns{UserID: authorUserID}}
	if arg.Input.Body != nil {
		comment.Body = *arg.Input.Body
	}

	campaign, err := dbCampaigns{}.Create(ctx, v, comment)
	if err != nil {
		return nil, err
	}

	if arg.Input.Rules != nil {
		for _, newRule := range *arg.Input.Rules {
			if _, err := (rules.DBRules{}).Create(ctx, &rules.DBRule{
				Container:   rules.RuleContainer{Campaign: campaign.ID},
				Name:        newRule.Name,
				Description: newRule.Description,
				Definition:  newRule.Definition,
				CreatedAt:   campaign.CreatedAt,
				UpdatedAt:   campaign.UpdatedAt,
			}); err != nil {
				return nil, err
			}
		}
	}

	x := &rulesExecutor{input: arg.Input}
	if err := x.executeRules(ctx, campaign.ID, campaign.Name); err != nil {
		return nil, err
	}

	return newGQLCampaign(campaign), nil
}

func (GraphQLResolver) UpdateCampaign(ctx context.Context, arg *graphqlbackend.UpdateCampaignArgs) (graphqlbackend.Campaign, error) {
	l, err := campaignByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	campaign, err := dbCampaigns{}.Update(ctx, l.db.ID, dbCampaignUpdate{
		Name: arg.Input.Name,
		// TODO!(sqs): description, renamed to body but allow it to be updated here
	})
	if err != nil {
		return nil, err
	}
	return newGQLCampaign(campaign), nil
}

func (GraphQLResolver) ForceRefreshCampaign(ctx context.Context, arg *graphqlbackend.ForceRefreshCampaignArgs) (graphqlbackend.Campaign, error) {
	campaign, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}

	campaignThreads, err := dbCampaignsThreads{}.List(ctx, dbCampaignsThreadsListOptions{CampaignID: campaign.db.ID})
	if err != nil {
		return nil, err
	}
	for _, campaignThread := range campaignThreads {
		if err := threads.Refresh(ctx, campaignThread.Thread); err != nil {
			return nil, err
		}
	}
	return campaign, nil
}

func (GraphQLResolver) DeleteCampaign(ctx context.Context, arg *graphqlbackend.DeleteCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlCampaign, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}
	return nil, dbCampaigns{}.DeleteByID(ctx, gqlCampaign.db.ID)
}

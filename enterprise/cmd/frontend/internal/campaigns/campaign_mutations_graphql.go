package campaigns

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// TODO!(sqs): use DB transactions here

func (GraphQLResolver) CreateCampaign(ctx context.Context, arg *graphqlbackend.ExpCreateCampaignArgs) (graphqlbackend.Campaign, error) {
	data := &dbCampaign{
		Name:      arg.Input.Name,
		IsDraft:   arg.Input.Draft != nil && *arg.Input.Draft,
		StartDate: arg.Input.StartDate.TimeOrNil(),
		DueDate:   arg.Input.DueDate.TimeOrNil(),
	}
	b, err := json.Marshal(arg.Input.ExtensionData)
	if err != nil {
		return nil, err
	}
	data.ExtensionData = json.RawMessage(b)
	data.NamespaceUserID, data.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, arg.Input.Namespace)
	if err != nil {
		return nil, err
	}

	if arg.Input.Body != nil {
		data.Description = *arg.Input.Body
	}
	if arg.Input.WorkflowAsJSONCString != nil {
		data.WorkflowJSONC = string(*arg.Input.WorkflowAsJSONCString)
	}

	author, err := comments.CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	data.AuthorID = author.UserID

	campaign, err := dbCampaigns{}.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	if err := replaceAndExecuteCampaignRules(ctx, campaign, campaign.CreatedAt, &arg.Input.ExtensionData); err != nil {
		return nil, err
	}
	return newGQLCampaign(campaign), nil
}

func (GraphQLResolver) UpdateCampaign(ctx context.Context, arg *graphqlbackend.ExpUpdateCampaignArgs) (graphqlbackend.Campaign, error) {
	c, err := campaignByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}

	data := dbCampaignUpdate{
		Name:           arg.Input.Name,
		StartDate:      arg.Input.StartDate.TimeOrNil(),
		ClearStartDate: arg.Input.ClearStartDate != nil && *arg.Input.ClearStartDate,
		DueDate:        arg.Input.DueDate.TimeOrNil(),
		ClearDueDate:   arg.Input.ClearDueDate != nil && *arg.Input.ClearDueDate,
		// TODO!(sqs): description, renamed to body but allow it to be updated here
	}
	if arg.Input.ExtensionData != nil {
		b, err := json.Marshal(arg.Input.ExtensionData)
		if err != nil {
			return nil, err
		}
		data.ExtensionData = json.RawMessage(b)
	}

	campaign, err := dbCampaigns{}.Update(ctx, c.db.ID, data)
	if err != nil {
		return nil, err
	}

	if arg.Input.WorkflowAsJSONCString != nil {
		// TODO!(sqs): actually perform a delta-update of threads here instead of deleting them all and
		// then recreating them.
		if _, err := dbconn.Global.ExecContext(ctx, `DELETE FROM threads USING exp_campaigns_threads WHERE threads.id=exp_campaigns_threads.thread_id AND exp_campaigns_threads.campaign_id=$1`, campaign.ID); err != nil {
			return nil, err
		}

		if err := replaceAndExecuteCampaignRules(ctx, campaign, time.Now(), arg.Input.ExtensionData); err != nil {
			return nil, err
		}
	}
	return newGQLCampaign(campaign), nil
}

func replaceAndExecuteCampaignRules(ctx context.Context, campaign *dbCampaign, ruleDate time.Time, extensionData *graphqlbackend.CampaignExtensionData) error {
	x := &rulesExecutor{
		campaign: ruleExecutorCampaignInfo{
			id:      campaign.ID,
			name:    campaign.Name,
			body:    campaign.Description,
			isDraft: campaign.IsDraft,
		},
		extensionData: extensionData,
	}
	return x.executeRules(ctx)
}

func (GraphQLResolver) PublishDraftCampaign(ctx context.Context, arg *graphqlbackend.PublishDraftCampaignArgs) (graphqlbackend.Campaign, error) {
	l, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}
	if !l.db.IsDraft {
		return nil, errors.New("campaign is not a draft campaign")
	}

	ts, err := l.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	for _, thread := range ts {
		if thread.Common().IsPendingExternalCreation() {
			if err := threads.PublishThreadToExternalService(ctx, thread.Thread); err != nil {
				return nil, err
			}
		}
	}

	tmp := false
	campaign, err := dbCampaigns{}.Update(ctx, l.db.ID, dbCampaignUpdate{
		IsDraft: &tmp,
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
	body, err := campaign.Body(ctx)
	if err != nil {
		return nil, err
	}

	x := &rulesExecutor{
		campaign: ruleExecutorCampaignInfo{
			id:      campaign.db.ID,
			name:    campaign.db.Name,
			body:    body,
			isDraft: campaign.db.IsDraft,
		},
		extensionData: &arg.ExtensionData,
	}
	if err := x.executeRules(ctx); err != nil {
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

func (GraphQLResolver) DeleteCampaign(ctx context.Context, arg *graphqlbackend.ExpDeleteCampaignArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlCampaign, err := campaignByID(ctx, arg.Campaign)
	if err != nil {
		return nil, err
	}
	return nil, dbCampaigns{}.DeleteByID(ctx, gqlCampaign.db.ID)
}

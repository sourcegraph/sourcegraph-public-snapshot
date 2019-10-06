package campaigns

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rules"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

// TODO!(sqs): use DB transactions here

func (GraphQLResolver) CreateCampaign(ctx context.Context, arg *graphqlbackend.ExpCreateCampaignArgs) (graphqlbackend.Campaign, error) {
	data := &dbCampaign{
		Name:      arg.Input.Name,
		IsDraft:   arg.Input.Draft != nil && *arg.Input.Draft,
		StartDate: arg.Input.StartDate.TimeOrNil(),
		DueDate:   arg.Input.DueDate.TimeOrNil(),
		// TODO!(sqs): description, renamed to body but allow it to be updated here
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

	author, err := comments.CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	comment := commentobjectdb.DBObjectCommentFields{Author: author}
	if arg.Input.Body != nil {
		comment.Body = *arg.Input.Body
	}

	campaign, err := dbCampaigns{}.Create(ctx, data, comment)
	if err != nil {
		return nil, err
	}
	if arg.Input.Rules != nil {
		if err := replaceAndExecuteCampaignRules(ctx, campaign, &comment.Body, *arg.Input.Rules, campaign.CreatedAt, &arg.Input.ExtensionData); err != nil {
			return nil, err
		}
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

	if arg.Input.Rules != nil {
		// TODO!(sqs): actually perform a delta-update of threads here instead of deleting them all and
		// then recreating them.
		if _, err := dbconn.Global.ExecContext(ctx, `DELETE FROM threads USING exp_campaigns_threads WHERE threads.id=exp_campaigns_threads.thread_id AND exp_campaigns_threads.campaign_id=$1`, campaign.ID); err != nil {
			return nil, err
		}

		if err := replaceAndExecuteCampaignRules(ctx, campaign, nil, *arg.Input.Rules, time.Now(), arg.Input.ExtensionData); err != nil {
			return nil, err
		}
	}
	return newGQLCampaign(campaign), nil
}

func replaceAndExecuteCampaignRules(ctx context.Context, campaign *dbCampaign, campaignBodyIfKnown *string, newRules []graphqlbackend.NewRuleInput, ruleDate time.Time, extensionData *graphqlbackend.CampaignExtensionData) error {
	var campaignBody string
	if campaignBodyIfKnown != nil {
		campaignBody = *campaignBodyIfKnown
	} else {
		comment, err := comments.DBGetByID(ctx, campaign.PrimaryCommentID)
		if err != nil {
			return err
		}
		campaignBody = comment.Body
	}

	if len(newRules) > 0 && extensionData == nil {
		return errors.New("executing campaign rules requires extensionData")
	}
	ruleContainer := rules.RuleContainer{Campaign: campaign.ID}

	// Remove existing rules.
	//
	// TODO(sqs): Updating rules in-place would be nice.
	if err := (rules.DBRules{}).DeleteByContainer(ctx, ruleContainer); err != nil {
		return err
	}

	// Add rules.
	for _, newRule := range newRules {
		data := &rules.DBRule{
			Container:   ruleContainer,
			Name:        newRule.Name,
			Description: newRule.Description,
			Definition:  string(newRule.Definition),
			CreatedAt:   ruleDate,
			UpdatedAt:   ruleDate,
		}
		if newRule.Template != nil {
			data.TemplateID = &newRule.Template.Template
			var err error
			data.TemplateContext, err = newRule.Template.ContextJSONCString()
			if err != nil {
				return err
			}
		}

		if _, err := (rules.DBRules{}).Create(ctx, data); err != nil {
			return err
		}
	}
	x := &rulesExecutor{
		campaign: ruleExecutorCampaignInfo{
			id:      campaign.ID,
			name:    campaign.Name,
			body:    campaignBody,
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

	ruleConnection, err := campaign.Rules(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return nil, err
	}
	rules, err := ruleConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	ruleData := make([]graphqlbackend.NewRuleInput, len(rules))
	for i, rule := range rules {
		def, err := jsonc.Parse(string(rule.Definition().Raw()))
		if err != nil {
			return nil, err
		}
		ruleData[i] = graphqlbackend.NewRuleInput{
			Name:        rule.Name(),
			Description: rule.Description(),
			Definition:  graphqlbackend.JSONCString(def),
		}
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

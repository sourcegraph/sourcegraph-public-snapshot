package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
)

const (
	eventTypeAddThreadToCampaign      = "AddThreadToCampaign"
	eventTypeRemoveThreadFromCampaign = "RemoveThreadFromCampaign"
)

func init() {
	events.Register(eventTypeAddThreadToCampaign, func(ctx context.Context, common graphqlbackend.EventCommon, data *events.EventData, toEvent *graphqlbackend.ToEvent) error {
		campaign, err := campaignByDBID(ctx, data.CampaignID)
		if err != nil {
			return err
		}
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.ThreadID)
		if err != nil {
			return err
		}
		toEvent.AddThreadToCampaignEvent = &graphqlbackend.AddRemoveThreadToFromCampaignEvent{
			EventCommon: common,
			Campaign:    campaign,
			Thread:      thread,
		}
		return nil
	})
}

package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
)

const (
	eventTypeAddThreadToCampaign      = "AddThreadToCampaign"
	eventTypeRemoveThreadFromCampaign = "RemoveThreadFromCampaign"
)

func init() {
	events.Register(eventTypeAddThreadToCampaign, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		campaign, err := campaignByDBID(ctx, data.Campaign)
		if err != nil {
			return err
		}
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.Thread)
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

func (v *gqlCampaign) TimelineItems(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.EventConnection, error) {
	// TODO!(sqs): filter by only events related to this campaign, and include events on its threads
	return events.GetEventConnection(ctx, &graphqlbackend.EventsArgs{ConnectionArgs: *arg}) // TODO!(sqs): support Since arg field
}

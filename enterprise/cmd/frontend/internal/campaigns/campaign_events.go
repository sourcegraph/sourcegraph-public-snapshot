package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

const (
	eventTypeAddThreadToCampaign      events.Type = "AddThreadToCampaign"
	eventTypeRemoveThreadFromCampaign             = "RemoveThreadFromCampaign"
)

func init() {
	events.Register(eventTypeAddThreadToCampaign, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		campaign, err := campaignByDBID(ctx, data.Campaign)
		if err != nil {
			return err
		}
		thread, err := graphqlbackend.ThreadByID(ctx, graphqlbackend.MarshalThreadID(data.Thread))
		if err != nil {
			return err
		}
		toEvent.AddThreadToCampaignEvent = &graphqlbackend.AddRemoveThreadToFromCampaignEvent{
			EventCommon: common,
			Campaign_:   campaign,
			Thread_:     thread,
		}
		return nil
	})
}

func (v *gqlCampaign) TimelineItems(ctx context.Context, arg *graphqlbackend.EventConnectionCommonArgs) (graphqlbackend.EventConnection, error) {
	return events.GetEventConnection(ctx,
		arg,
		events.Objects{Campaign: v.db.ID},
	)
}

package campaigns

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

const (
	eventTypeAddThreadToCampaign      events.Type = "AddThreadToCampaign"
	eventTypeRemoveThreadFromCampaign             = "RemoveThreadFromCampaign"
)

func init() {
	for _, eventType_ := range []events.Type{eventTypeAddThreadToCampaign, eventTypeRemoveThreadFromCampaign} {
		eventType := eventType_
		events.Register(eventType, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
			campaign, err := campaignByDBID(ctx, data.Campaign)
			if err != nil {
				return err
			}
			thread, err := graphqlbackend.ThreadByID(ctx, graphqlbackend.MarshalThreadID(data.Thread))
			if err != nil {
				return err
			}
			event := &graphqlbackend.AddRemoveThreadToFromCampaignEvent{
				EventCommon: common,
				Campaign_:   campaign,
				Thread_:     thread,
			}
			switch {
			case eventType == eventTypeAddThreadToCampaign:
				toEvent.AddThreadToCampaignEvent = event
			case eventType == eventTypeRemoveThreadFromCampaign:
				toEvent.RemoveThreadFromCampaignEvent = event
			default:
				panic("unexpected event type: " + eventType)
			}
			return nil
		})
	}
}

type eventsGetter interface {
	// getEvents returns a list of events related to the campaign that occurred before a given date.

	// In the list of events for a campaign, we include events on a thread in the campaign before
	// the thread was added to the campaign. This is so that you can use campaigns to track efforts
	// that you started before you created the campaign.
	getEvents(ctx context.Context, beforeDate time.Time, eventTypes []events.Type) ([]graphqlbackend.ToEvent, error)
}

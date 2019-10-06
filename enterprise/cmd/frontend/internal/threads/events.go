package threads

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

const (
	eventTypeCreateThread    events.Type = "CreateThread"
	eventTypeReview          events.Type = "Review"
	eventTypeReviewRequested events.Type = "ReviewRequested"
	eventTypeMergeThread     events.Type = "MergeThread"
	eventTypeCloseThread     events.Type = "CloseThread"
	eventTypeReopenThread    events.Type = "ReopenThread"
)

func init() {
	events.Register(eventTypeCreateThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.CreateThreadEvent = &graphqlbackend.CreateThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeReview, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		// TODO!(sqs): validate state
		var o struct {
			State graphqlbackend.ReviewState `json:"state"`
		}
		if err := json.Unmarshal(data.Data, &o); err != nil {
			return err
		}
		toEvent.ReviewEvent = &graphqlbackend.ReviewEvent{
			EventCommon: common,
			Thread_:     thread,
			State_:      o.State,
		}
		return nil
	})
	events.Register(eventTypeReviewRequested, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.RequestReviewEvent = &graphqlbackend.RequestReviewEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeMergeThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.MergeThreadEvent = &graphqlbackend.MergeThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeCloseThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.CloseThreadEvent = &graphqlbackend.CloseThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeReopenThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.ReopenThreadEvent = &graphqlbackend.ReopenThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
}

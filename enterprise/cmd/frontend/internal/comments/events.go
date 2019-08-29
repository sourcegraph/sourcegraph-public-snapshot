package comments

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

const EventTypeComment events.Type = "Comment"

func init() {
	events.Register(EventTypeComment, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		// TODO!(sqs): read comment data from event payload
		toEvent.CommentEvent = &graphqlbackend.CommentEvent{
			EventCommon: common,
			// TODO!(sqs): store comment data
		}
		return nil
	})
}

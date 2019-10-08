package comments

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
)

const EventTypeComment events.Type = "Comment"

func init() {
	events.Register(EventTypeComment, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		if data.Comment == 0 {
			return errors.New("invalid comment event with comment ID 0")
		}
		dbComment, err := internal.DBComments{}.GetByID(ctx, data.Comment)
		if err != nil {
			return err
		}
		comment, err := newGQLToComment(ctx, dbComment)
		if err != nil {
			return err
		}
		toEvent.CommentEvent = &graphqlbackend.CommentEvent{
			EventCommon: common,
			Comment_:    comment,
		}
		return nil
	})
}

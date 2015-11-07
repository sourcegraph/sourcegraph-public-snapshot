package discussionhooks

import (
	"github.com/AaronO/go-git-http"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
)

func init() {
	events.RegisterListener(&discussionListener{})
}

type discussionListener struct{}

func (g *discussionListener) Scopes() []string {
	return []string{"app:discussions"}
}

func (g *discussionListener) Start(ctx context.Context) {
	notifyCallback := func(id events.EventID, p notif.DiscussionPayload) {
		notifyDiscussionEvent(ctx, id, p)
	}

	events.Subscribe(notif.DiscussionCreateEvent, notifyCallback)
	events.Subscribe(notif.DiscussionCommentEvent, notifyCallback)
}

func notifyDiscussionEvent(ctx context.Context, id events.EventID, payload notif.DiscussionPayload) {
	cl := sourcegraph.NewClientFromContext(ctx)

	if payload.Discussion == nil {
		return
	}

	var recipients []*sourcegraph.UserSpec
	var actionType string
	var err error

	switch id {
	case notif.DiscussionCreateEvent:
		recipients, err = mdutil.Mentions(ctx, []byte(payload.Discussion.Description))
		if err != nil {
			return nil, err
		}
		actionType = "created"

	case notif.DiscussionCommentEvent:
		if payload.Comment == nil {
			return
		}
		if payload.Discussion.Author.UID != payload.Actor.UID {
			recipients = append(recipients, &payload.Discussion.Author)
		}
		ppl, err := mdutil.Mentions(ctx, []byte(payload.Comment.Body))
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, ppl...)
		actionType = "commented on"
	}

	// Send notification
	cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
		Actor:       payload.Actor,
		Recipients:  recipients,
		ActionType:  actionType,
		ObjectURL:   payload.URL,
		ObjectRepo:  payload.Repo,
		ObjectType:  "discussion",
		ObjectID:    payload.ID,
		ObjectTitle: payload.Title,
	})
}

package listeners

import (
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

func init() {
	events.RegisterListener(&discussionListener{})
}

type discussionListener struct{}

func (g *discussionListener) Scopes() []string {
	return []string{"app:discussions"}
}

func (g *discussionListener) Start(ctx context.Context) {
	notifyCallback := func(id events.EventID, p events.DiscussionPayload) {
		notifyDiscussionEvent(ctx, id, p)
	}

	events.Subscribe(events.DiscussionCreateEvent, notifyCallback)
	events.Subscribe(events.DiscussionCommentEvent, notifyCallback)
}

func notifyDiscussionEvent(ctx context.Context, id events.EventID, payload events.DiscussionPayload) {
	cl := sourcegraph.NewClientFromContext(ctx)

	if payload.Discussion == nil {
		return
	}

	var recipients []*sourcegraph.UserSpec
	var actionType string
	var err error

	switch id {
	case events.DiscussionCreateEvent:
		recipients, err = mdutil.Mentions(ctx, []byte(payload.Discussion.Description))
		if err != nil {
			log15.Warn("DiscussionHook: ignoring event", "event", id, "error", err)
			return
		}
		actionType = "created"

	case events.DiscussionCommentEvent:
		if payload.Comment == nil {
			return
		}
		if payload.Discussion.Author.UID != payload.Actor.UID {
			recipients = append(recipients, &payload.Discussion.Author)
		}
		ppl, err := mdutil.Mentions(ctx, []byte(payload.Comment.Body))
		if err != nil {
			log15.Warn("DiscussionHook: ignoring event", "event", id, "error", err)
			return
		}
		recipients = append(recipients, ppl...)
		actionType = "commented on"
	}

	// Send notification
	cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
		Actor:       &payload.Actor,
		Recipients:  recipients,
		ActionType:  actionType,
		ObjectURL:   payload.URL,
		ObjectRepo:  payload.Repo,
		ObjectType:  "discussion",
		ObjectID:    payload.ID,
		ObjectTitle: payload.Title,
	})
}

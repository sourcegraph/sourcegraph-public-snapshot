package changesets

import (
	"bytes"
	"fmt"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif"
	csnotif "src.sourcegraph.com/sourcegraph/platform/apps/changesets/notif"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

func init() {
	events.RegisterListener(&changesetListener{})
}

type changesetListener struct{}

func (g *changesetListener) Scopes() []string {
	return []string{"app:changes"}
}

func (g *changesetListener) Start(ctx context.Context) {
	gitCallback := func(id events.EventID, payload notif.GitPayload) {
		if couldAffectChangesets(id, payload) {
			updateAffectedChangesets(ctx, id, payload)
		}
	}

	events.Subscribe(notif.GitPushEvent, gitCallback)
	events.Subscribe(notif.GitDeleteEvent, gitCallback)

	notifyCallback := func(id events.EventID, payload csnotif.ChangesetPayload) {
		notifyChangesetEvent(ctx, id, payload)
	}

	events.Subscribe(csnotif.ChangesetCreateEvent, notifyCallback)
	events.Subscribe(csnotif.ChangesetReviewEvent, notifyCallback)
	events.Subscribe(csnotif.ChangesetUpdateEvent, notifyCallback)
	events.Subscribe(csnotif.ChangesetCloseEvent, notifyCallback)
}

func updateAffectedChangesets(ctx context.Context, id events.EventID, payload notif.GitPayload) {
	e := payload.Event
	cl := sourcegraph.NewClientFromContext(ctx)
	changesetEvents, err := cl.Changesets.UpdateAffected(ctx, &sourcegraph.ChangesetUpdateAffectedOp{
		Repo:   payload.Repo,
		Branch: e.Branch,
		Last:   e.Last,
		Commit: e.Commit,
	})
	if err != nil {
		log15.Warn("changesetHook: could not update changesets", "error", err)
	}

	for _, e := range changesetEvents.Events {
		op := e.Op
		cspayload := csnotif.ChangesetPayload{
			Actor:  payload.Actor,
			ID:     op.ID,
			Repo:   op.Repo.URI,
			Title:  op.Title,
			URL:    urlToChangeset(ctx, op.ID),
			Update: op,
		}
		if op.Close {
			events.Publish(csnotif.ChangesetCloseEvent, cspayload)
		} else {
			events.Publish(csnotif.ChangesetUpdateEvent, cspayload)
		}
	}
}

func notifyChangesetEvent(ctx context.Context, id events.EventID, payload csnotif.ChangesetPayload) {
	switch id {
	case csnotif.ChangesetCreateEvent:
		notifyCreation(ctx, payload)
	case csnotif.ChangesetReviewEvent:
		notifyReview(ctx, payload)
	default:
		// TODO: implement notifications for changeset update and close events
		return
	}
}

// notifyCreation creates a slack notification that a changeset was created. It
// also notifies users mentioned in the description of the changeset.
func notifyCreation(ctx context.Context, payload csnotif.ChangesetPayload) {
	if payload.Changeset == nil {
		return
	}

	cl := sourcegraph.NewClientFromContext(ctx)

	// Build list of recipients
	recipients, err := mdutil.Mentions(ctx, []byte(payload.Changeset.Description))
	if err != nil {
		return
	}

	// Send notification
	cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
		Actor:         &payload.Actor,
		Recipients:    recipients,
		ActionType:    "created",
		ObjectURL:     payload.URL,
		ObjectRepo:    payload.Repo,
		ObjectType:    "changeset",
		ObjectID:      payload.ID,
		ObjectTitle:   payload.Title,
		ActionContent: payload.Changeset.Description,
	})
}

// notifyReview creates a slack notification that a changeset was reviewed. It
// also notifies any users potentially mentioned in the review.
func notifyReview(ctx context.Context, payload csnotif.ChangesetPayload) {
	if payload.Changeset == nil || payload.Review == nil {
		return
	}

	cl := sourcegraph.NewClientFromContext(ctx)

	// Build list of recipients
	recipients, err := mdutil.Mentions(ctx, []byte(payload.Review.Body))
	if err != nil {
		return
	}
	for _, c := range payload.Review.Comments {
		mentions, err := mdutil.Mentions(ctx, []byte(c.Body))
		if err != nil {
			return
		}
		recipients = append(recipients, mentions...)
	}
	recipients = append(recipients, &payload.Changeset.Author)

	// Send notification
	msg := bytes.NewBufferString(payload.Review.Body)
	for _, c := range payload.Review.Comments {
		msg.WriteString(fmt.Sprintf("\n*%s:%d* - %s", c.Filename, c.LineNumber, c.Body))
	}
	cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
		Actor:         &payload.Actor,
		Recipients:    recipients,
		ActionType:    "reviewed",
		ObjectURL:     payload.URL,
		ObjectRepo:    payload.Repo,
		ObjectType:    "changeset",
		ObjectID:      payload.ID,
		ObjectTitle:   payload.Title,
		ActionContent: msg.String(),
	})
}

// couldAffectChangesets returns true if the event was error-free
// and is a GitPushEvent or GitDeleteEvent.
func couldAffectChangesets(id events.EventID, p notif.GitPayload) bool {
	if !(id == notif.GitPushEvent || id == notif.GitDeleteEvent) {
		return false
	}
	e := p.Event
	if e.Error != nil || e.Branch == "" || !commitsValid(e.Commit, e.Last) {
		return false
	}
	return true
}

// commitsValid returns true if all commits in the paramters are exactly 40
// characters long.
func commitsValid(commits ...string) bool {
	for _, c := range commits {
		if len(c) != 40 {
			return false
		}
	}
	return true
}

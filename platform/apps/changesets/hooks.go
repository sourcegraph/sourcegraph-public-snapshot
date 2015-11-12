package changesets

import (
	"bytes"
	"fmt"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/events"
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
	gitCallback := func(id events.EventID, payload events.GitPayload) {
		if couldAffectChangesets(id, payload) {
			updateAffectedChangesets(ctx, id, payload)
		}
	}

	events.Subscribe(events.GitPushEvent, gitCallback)
	events.Subscribe(events.GitDeleteBranchEvent, gitCallback)

	notifyCallback := func(id events.EventID, payload events.ChangesetPayload) {
		notifyChangesetEvent(ctx, id, payload)
	}

	events.Subscribe(events.ChangesetCreateEvent, notifyCallback)
	events.Subscribe(events.ChangesetReviewEvent, notifyCallback)
	events.Subscribe(events.ChangesetUpdateEvent, notifyCallback)
	events.Subscribe(events.ChangesetCloseEvent, notifyCallback)
	events.Subscribe(events.ChangesetMergeEvent, notifyCallback)
}

func updateAffectedChangesets(ctx context.Context, id events.EventID, payload events.GitPayload) {
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
		return
	}

	for _, e := range changesetEvents.Events {
		op := e.Op
		cspayload := events.ChangesetPayload{
			Actor:  payload.Actor,
			ID:     op.ID,
			Repo:   op.Repo.URI,
			Title:  op.Title,
			Update: op,
		}
		if op.Merged {
			events.Publish(events.ChangesetMergeEvent, cspayload)
		} else if op.Close {
			events.Publish(events.ChangesetCloseEvent, cspayload)
		} else {
			events.Publish(events.ChangesetUpdateEvent, cspayload)
		}
	}
}

func notifyChangesetEvent(ctx context.Context, id events.EventID, payload events.ChangesetPayload) {
	if payload.URL == "" {
		changesetURL, err := urlToRepoChangeset(payload.Repo, payload.ID)
		if err == nil {
			payload.URL = conf.AppURL(ctx).ResolveReference(changesetURL).String()
		}
	}

	switch id {
	case events.ChangesetCreateEvent:
		notifyCreation(ctx, payload)
	case events.ChangesetReviewEvent:
		notifyReview(ctx, payload)
	case events.ChangesetUpdateEvent, events.ChangesetCloseEvent, events.ChangesetMergeEvent:
		notifyUpdate(ctx, id, payload)
	default:
		log15.Warn("changesetListener: unknown event id", "id", id)
		return
	}
}

// notifyCreation creates a slack notification that a changeset was created. It
// also notifies users mentioned in the description of the changeset.
func notifyCreation(ctx context.Context, payload events.ChangesetPayload) {
	if payload.Changeset == nil {
		log15.Warn("changesetListener: no changeset in context", "payload", payload)
		return
	}

	cl := sourcegraph.NewClientFromContext(ctx)

	// Build list of recipients
	recipients, err := mdutil.Mentions(ctx, []byte(payload.Changeset.Description))
	if err != nil {
		log15.Warn("changesetListener: error parsing mentions from changeset description", "error", err)
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
func notifyReview(ctx context.Context, payload events.ChangesetPayload) {
	if payload.Changeset == nil || payload.Review == nil {
		log15.Warn("changesetListener: no changeset or review in context", "payload", payload)
		return
	}

	cl := sourcegraph.NewClientFromContext(ctx)

	// Build list of recipients
	recipients, err := mdutil.Mentions(ctx, []byte(payload.Review.Body))
	if err != nil {
		log15.Warn("changesetListener: error parsing mentions from changeset description", "error", err)
		return
	}
	for _, c := range payload.Review.Comments {
		mentions, err := mdutil.Mentions(ctx, []byte(c.Body))
		if err != nil {
			log15.Warn("changesetListener: error parsing mentions from changeset review", "error", err)
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

// notifyUpdate creates a slack notification that a changeset was updated, closed or merged.
func notifyUpdate(ctx context.Context, id events.EventID, payload events.ChangesetPayload) {
	cl := sourcegraph.NewClientFromContext(ctx)

	var actionType string
	switch id {
	case events.ChangesetUpdateEvent:
		// don't send notifications about changeset updates
		return
	case events.ChangesetCloseEvent:
		actionType = "closed"
	case events.ChangesetMergeEvent:
		actionType = "merged"
	default:
		log15.Warn("changesetListener: unknown event id", "id", id)
		return
	}

	// Send notification
	cl.Notify.GenericEvent(ctx, &sourcegraph.NotifyGenericEvent{
		Actor:       &payload.Actor,
		ActionType:  actionType,
		ObjectURL:   payload.URL,
		ObjectRepo:  payload.Repo,
		ObjectType:  "changeset",
		ObjectID:    payload.ID,
		ObjectTitle: payload.Title,
	})
}

// couldAffectChangesets returns true if the event was error-free
// and is a GitPushEvent or GitDeleteBranchEvent.
func couldAffectChangesets(id events.EventID, p events.GitPayload) bool {
	if !(id == events.GitPushEvent || id == events.GitDeleteBranchEvent) {
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

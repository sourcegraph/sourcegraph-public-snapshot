package changesets

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/russross/blackfriday"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	githttp "github.com/AaronO/go-git-http"
	notif "src.sourcegraph.com/apps/notifications/notifications"
	"src.sourcegraph.com/apps/tracker/issues"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/notifications"

	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
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
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("changesetHook: could not create client", "error", err)
		return
	}
	changesetEvents, err := cl.Changesets.UpdateAffected(ctx, &sourcegraph.ChangesetUpdateAffectedOp{
		Repo:      payload.Repo,
		Branch:    e.Branch,
		Last:      e.Last,
		Commit:    e.Commit,
		ForcePush: e.Type == githttp.PUSH_FORCE,
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
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("changesetListener error", "error", err)
	}
	if payload.Changeset == nil {
		cs, err := cl.Changesets.Get(ctx, &sourcegraph.ChangesetSpec{
			Repo: sourcegraph.RepoSpec{URI: payload.Repo},
			ID:   payload.ID,
		})
		if err != nil {
			log15.Warn("changesetListener: could not fetch changeset", "repo", payload.Repo, "id", payload.ID, "error", err)
			return
		}
		payload.Changeset = cs
	}

	if payload.URL == "" {
		changesetURL, err := urlToRepoChangeset(payload.Repo, payload.ID)
		if err == nil {
			payload.URL = conf.AppURL(ctx).ResolveReference(changesetURL).String()
		}
	}

	if payload.Title == "" {
		payload.Title = payload.Changeset.Title
	}

	if flags.JiraURL != "" {
		jiraOnChangesetUpdate(ctx, payload.Changeset)
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

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("changesetListener: error creating client", "error", err)
	}

	// Build list of recipients
	recipients, err := mdutil.Mentions(ctx, []byte(payload.Changeset.Description))
	if err != nil {
		log15.Warn("changesetListener: error parsing mentions from changeset description", "error", err)
		return
	}

	n := &sourcegraph.NotifyGenericEvent{
		Actor:         &payload.Actor,
		Recipients:    recipients,
		ActionType:    "created",
		ObjectURL:     payload.URL,
		ObjectRepo:    payload.Repo,
		ObjectType:    "changeset",
		ObjectID:      payload.ID,
		ObjectTitle:   payload.Title,
		ActionContent: payload.Changeset.Description,
		EmailHTML:     string(blackfriday.MarkdownCommon([]byte(payload.Changeset.Description))),
	}

	// Send notification
	// TODO: Unified API for notifications
	cl.Notify.GenericEvent(ctx, n)
	notificationCenter(ctx, n)
}

// notifyReview creates a slack notification that a changeset was reviewed. It
// also notifies any users potentially mentioned in the review.
func notifyReview(ctx context.Context, payload events.ChangesetPayload) {
	if payload.Changeset == nil || payload.Review == nil {
		log15.Warn("changesetListener: no changeset or review in context", "payload", payload)
		return
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("changesetListener: error creating client", "error", err)
	}

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

	msg := bytes.NewBufferString(payload.Review.Body)
	for _, c := range payload.Review.Comments {
		msg.WriteString(fmt.Sprintf("\n\n- *%s:%d* - %s", c.Filename, c.LineNumber, c.Body))
	}
	actionContent := msg.String()
	n := &sourcegraph.NotifyGenericEvent{
		Actor:         &payload.Actor,
		Recipients:    recipients,
		ActionType:    "reviewed",
		ObjectURL:     payload.URL,
		ObjectRepo:    payload.Repo,
		ObjectType:    "changeset",
		ObjectID:      payload.ID,
		ObjectTitle:   payload.Title,
		ActionContent: actionContent,
		EmailHTML:     string(blackfriday.MarkdownCommon([]byte(actionContent))),
	}

	// Send notification
	// TODO: Unified API for notifications
	cl.Notify.GenericEvent(ctx, n)
	notificationCenter(ctx, n)
}

// notifyUpdate creates a slack notification that a changeset was updated, closed or merged.
func notifyUpdate(ctx context.Context, id events.EventID, payload events.ChangesetPayload) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("changesetListener: error creating client", "error", err)
	}

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

	n := &sourcegraph.NotifyGenericEvent{
		Actor:       &payload.Actor,
		ActionType:  actionType,
		ObjectURL:   payload.URL,
		ObjectRepo:  payload.Repo,
		ObjectType:  "changeset",
		ObjectID:    payload.ID,
		ObjectTitle: payload.Title,
	}

	// Send notification
	// TODO: Unified API for notifications
	cl.Notify.GenericEvent(ctx, n)
	notificationCenter(ctx, n)
}

func notificationCenter(ctx context.Context, e *sourcegraph.NotifyGenericEvent) {
	if notifications.Service == nil {
		return
	}
	subscribers := []issues.UserSpec{issues.UserSpec{ID: uint64(e.Actor.UID), Domain: e.Actor.Domain}}
	if e.Recipients != nil {
		for _, u := range e.Recipients {
			subscribers = append(subscribers, issues.UserSpec{ID: uint64(u.UID), Domain: u.Domain})
		}
	}
	// HACK(keegancsmith) Notification API expects the user to be set in
	// the context like we have in HTTP requests. This context is from the
	// event bus. Fake it
	ctx = handlerutil.WithUser(ctx, e.Actor)
	notifications.Service.Subscribe(ctx, appID, issues.RepoSpec{URI: e.ObjectRepo}, uint64(e.ObjectID), subscribers)
	notifications.Service.Notify(ctx, appID, issues.RepoSpec{URI: e.ObjectRepo}, uint64(e.ObjectID), notif.Notification{
		Title:     e.ObjectTitle,
		Icon:      "git-pull-request",
		UpdatedAt: time.Now(),
		HTMLURL:   template.URL(e.ObjectURL),
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

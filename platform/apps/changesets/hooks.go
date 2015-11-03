package changesets

import (
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/events/githooks"
	"src.sourcegraph.com/sourcegraph/notif"
)

func init() {
	events.RegisterListener(&changesetHookListener{})
}

type changesetHookListener struct{}

func (g *changesetHookListener) Scopes() []string {
	return []string{"app:changes"}
}

func (g *changesetHookListener) Start(ctx context.Context) {
	callback := func(id events.EventID, p githooks.Payload) {
		if !couldAffectChangesets(id, p) {
			return
		}
		e := p.Event
		cl := sourcegraph.NewClientFromContext(ctx)
		changesetEvents, err := cl.Changesets.UpdateAffected(ctx, &sourcegraph.ChangesetUpdateAffectedOp{
			Repo:   p.Repo,
			Branch: e.Branch,
			Last:   e.Last,
			Commit: e.Commit,
		})
		if err != nil {
			log15.Warn("changesetHook: could not update changesets", "error", err)
		}

		userSpec := sourcegraph.UserSpec{
			UID:    int32(p.CtxActor.UID),
			Login:  p.CtxActor.Login,
			Domain: p.CtxActor.Domain,
		}

		for _, e := range changesetEvents.Events {
			op := e.Op
			payload := notif.ChangesetPayload{
				UserSpec: userSpec,
				ID:       op.ID,
				Repo:     op.Repo.URI,
				Title:    op.Title,
				URL:      urlToChangeset(ctx, op.ID),
				Update:   op,
			}
			if op.Close {
				events.Publish(notif.ChangesetCloseEvent, payload)
			} else {
				events.Publish(notif.ChangesetUpdateEvent, payload)
			}
		}
	}

	events.Subscribe(githooks.GitPushEvent, callback)
	events.Subscribe(githooks.GitDeleteEvent, callback)
}

// couldAffectChangesets returns true if the event was error-free
// and is a GitPushEvent or GitDeleteEvent.
func couldAffectChangesets(id events.EventID, p githooks.Payload) bool {
	if !(id == githooks.GitPushEvent || id == githooks.GitDeleteEvent) {
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

package changesets

import (
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/notif/githooks"
)

func init() {
	events.Listeners = append(events.Listeners, &changesetHookListener{})
}

type changesetHookListener struct{}

func (g *changesetHookListener) Scopes() []string {
	return []string{"app:changes"}
}

func (g *changesetHookListener) Start(ctx context.Context) {
	callback := func(p githooks.Payload) {
		if !couldAffectChangesets(p) {
			return
		}
		e := p.Event
		cl := sourcegraph.NewClientFromContext(ctx)
		_, err := cl.Changesets.UpdateAffected(ctx, &sourcegraph.ChangesetUpdateAffectedOp{
			Repo:   p.Repo,
			Branch: e.Branch,
			Last:   e.Last,
			Commit: e.Commit,
		})
		if err != nil {
			log15.Warn("changesetHook: could not update changesets", "error", err)
		}
	}

	events.Subscribe(githooks.GitPushEvent, callback)
	events.Subscribe(githooks.GitDeleteEvent, callback)
}

// couldAffectChangesets returns true if the event was error-free
// and is a GitPushEvent or GitDeleteEvent.
func couldAffectChangesets(p githooks.Payload) bool {
	e := p.Event
	if e.Error != nil || e.Branch == "" || !commitsValid(e.Commit, e.Last) {
		return false
	}
	if !(p.Type == githooks.GitPushEvent || p.Type == githooks.GitDeleteEvent) {
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

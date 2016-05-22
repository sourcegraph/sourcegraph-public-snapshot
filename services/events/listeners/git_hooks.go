package listeners

import (
	"fmt"
	"strings"

	"github.com/AaronO/go-git-http"
	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	appconf "sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/textutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/events"
)

func init() {
	events.RegisterListener(&gitHookListener{})
}

type gitHookListener struct{}

func (g *gitHookListener) Scopes() []string {
	return []string{"internal:githooks"}
}

func (g *gitHookListener) Start(ctx context.Context) {
	if !appconf.Flags.DisableGitNotify {
		notifyCallback := func(id events.EventID, p events.GitPayload) {
			notifyGitEvent(ctx, id, p)
		}
		events.Subscribe(events.GitPushEvent, notifyCallback)
		events.Subscribe(events.GitCreateBranchEvent, notifyCallback)
		events.Subscribe(events.GitDeleteBranchEvent, notifyCallback)
	}

	buildCallback := func(id events.EventID, p events.GitPayload) {
		buildHook(ctx, id, p)
	}
	events.Subscribe(events.GitPushEvent, buildCallback)
	events.Subscribe(events.GitCreateBranchEvent, buildCallback)

	inventoryCallback := func(id events.EventID, p events.GitPayload) {
		inventoryHook(ctx, id, p)
	}
	events.Subscribe(events.GitPushEvent, inventoryCallback)
	events.Subscribe(events.GitCreateBranchEvent, inventoryCallback)
}

func notifyGitEvent(ctx context.Context, id events.EventID, payload events.GitPayload) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Warn("postPushHook error", "error", err)
	}

	repo, err := cl.Repos.Get(ctx, &payload.Repo)
	if err != nil {
		log15.Warn("postPushHook error fetching repo", "repo", payload.Repo.URI, "error", err)
	}
	// Don't emit notifications for mirror repositories.
	if repo.Mirror {
		return
	}

	event := payload.Event
	branchURL := router.Rel.URLToRepoRev(repo.URI, event.Branch)
	if err != nil {
		log15.Warn("postPushHook: error resolving branch URL", "repo", repo.URI, "branch", event.Branch, "error", err)
		return
	}

	absBranchURL := conf.AppURL(ctx).ResolveReference(branchURL).String()
	notifyEvent := sourcegraph.NotifyGenericEvent{
		Actor:      &payload.Actor,
		ObjectURL:  absBranchURL,
		ObjectRepo: repo.URI + "@" + event.Branch,
		NoEmail:    true,
	}

	if id != events.GitPushEvent {
		switch id {
		case events.GitCreateBranchEvent:
			notifyEvent.ActionType = "created the branch"
		case events.GitDeleteBranchEvent:
			notifyEvent.ActionType = "deleted the branch"
		default:
			log15.Warn("postPushHook: unknown event id", "id", id, "repo", repo.URI, "branch", event.Branch)
			return
		}

		cl.Notify.GenericEvent(ctx, &notifyEvent)
		return
	}

	// See how many commits were pushed.
	commits, err := cl.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: repo.RepoSpec(),
		Opt: &sourcegraph.RepoListCommitsOptions{
			Head:        event.Commit,
			Base:        event.Last,
			ListOptions: sourcegraph.ListOptions{PerPage: 1000},
		},
	})
	if err != nil {
		log15.Warn("postPushHook: error fetching push commits", "error", err)
		commits = &sourcegraph.CommitList{}
	}

	var commitsNoun string
	if len(commits.Commits) == 1 {
		commitsNoun = "commit"
	} else {
		commitsNoun = "commits"
	}
	var commitMessages []string
	for i, c := range commits.Commits {
		if i > 10 {
			break
		}
		commitMessages = append(commitMessages, fmt.Sprintf("%s: %s",
			c.ID[:6],
			textutil.ShortCommitMessage(80, c.Message),
		))
	}

	notifyEvent.ActionType = fmt.Sprintf("pushed *%d %s* to", len(commits.Commits), commitsNoun)
	notifyEvent.ActionContent = strings.Join(commitMessages, "\n")
	cl.Notify.GenericEvent(ctx, &notifyEvent)
}

func buildHook(ctx context.Context, id events.EventID, payload events.GitPayload) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("postPushHook: failed to create build", "err", err)
		return
	}
	repo := payload.Repo
	event := payload.Event

	if payload.IgnoreBuild {
		return
	}

	if event.Type == githttp.PUSH || event.Type == githttp.PUSH_FORCE || event.Type == githttp.TAG {
		_, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{
			Repo:     repo,
			CommitID: event.Commit,
			Tag:      event.Tag,
			Branch:   event.Branch,
			Config:   sourcegraph.BuildConfig{Queue: true},
		})
		if err != nil {
			log15.Warn("postPushHook: failed to create build", "err", err, "repo", repo.URI, "commit", event.Commit, "branch", event.Branch, "tag", event.Tag)
			return
		}
		log15.Debug("postPushHook: build created", "repo", repo.URI, "branch", event.Branch, "tag", event.Tag, "commit", event.Commit, "event type", event.Type)
	}
}

// inventoryHook triggers a Repos.GetInventory call that computes the
// repo's inventory and caches it in a commit status (and saves it to
// the repo's Language field for default branch pushes). Then it is
// available immediately for future callers (which generally expect
// that operation to be fast).
func inventoryHook(ctx context.Context, id events.EventID, payload events.GitPayload) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("inventoryHook error", "err", err)
	}
	event := payload.Event
	if event.Type == githttp.PUSH || event.Type == githttp.PUSH_FORCE {
		repoRev := &sourcegraph.RepoRevSpec{RepoSpec: payload.Repo, CommitID: event.Commit}
		// Trigger a call to Repos.GetInventory so the inventory is
		// cached for subsequent calls.
		inv, err := cl.Repos.GetInventory(ctx, repoRev)
		if err != nil {
			log15.Warn("inventoryHook: call to Repos.GetInventory failed", "err", err, "repoRev", repoRev)
			return
		}

		// If this push is to the default branch, update the repo's
		// Language field with the primary language.
		repo, err := cl.Repos.Get(ctx, &repoRev.RepoSpec)
		if err != nil {
			log15.Warn("inventoryHook: call to Repos.Get failed", "err", err, "repoRev", repoRev)
			return
		}
		if event.Branch == repo.DefaultBranch {
			lang := inv.PrimaryProgrammingLanguage()
			if _, err := cl.Repos.Update(ctx, &sourcegraph.ReposUpdateOp{Repo: repo.RepoSpec(), Language: lang}); err != nil {
				log15.Warn("inventoryHook: call to Repos.Update to set language failed", "err", err, "repoRev", repoRev, "language", lang)
			}
		}
	}
}

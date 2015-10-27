package githooks

import (
	"fmt"
	"log"
	"strings"

	"github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/ext/slack"
	"src.sourcegraph.com/sourcegraph/util/textutil"
)

const GitPushEvent events.EventID = "git.push"
const GitCreateEvent events.EventID = "git.create"
const GitDeleteEvent events.EventID = "git.delete"

type Payload struct {
	Type            events.EventID
	CtxActor        authpkg.Actor
	Repo            sourcegraph.RepoSpec
	ContentEncoding string
	Event           githttp.Event
}

type gitHookListener struct {
	Ctx context.Context

	SlackHook      func(Payload)
	BuildHook      func(Payload)
	ChangesetsHook func(Payload)
}

func (g *gitHookListener) Init(ctx context.Context) {
	g.Ctx = ctx

	g.SlackHook = func(p Payload) {
		slackContributionsHook(g, p)
	}
	g.BuildHook = func(p Payload) {
		buildHook(g, p)
	}
	// TODO(pararth): add back changesets_hook
	// g.ChangesetsHook = func(p Payload) {
	// 	if p.Event.Error == nil && p.Event.Branch != "" {
	// 		changesetsHook(g, p)
	// 	}
	// }

	events.Subscribe(GitPushEvent, g.SlackHook)
	events.Subscribe(GitCreateEvent, g.SlackHook)
	events.Subscribe(GitDeleteEvent, g.SlackHook)

	events.Subscribe(GitPushEvent, g.BuildHook)

	// events.Subscribe(GitPushEvent, g.ChangesetHook)
}

var Listener *gitHookListener = &gitHookListener{}

func slackContributionsHook(gitListener *gitHookListener, payload Payload) {
	ctx := gitListener.Ctx
	cl := sourcegraph.NewClientFromContext(ctx)
	userStr, err := getUserDisplayName(cl, ctx, payload.CtxActor)
	if err != nil {
		log.Printf("postPushHook: error getting user: %s", err)
		return
	}

	repo := payload.Repo
	event := payload.Event
	branchURL, err := router.Rel.URLToRepoRev(repo.URI, event.Branch)
	if err != nil {
		log.Printf("postPushHook: error resolving branch URL for repo %s and branch %s: %v", repo.URI, event.Branch, err)
		return
	}

	absBranchURL := conf.AppURL(ctx).ResolveReference(branchURL).String()

	if payload.Type == GitCreateEvent {
		msg := fmt.Sprintf("*%s* created the branch <%s|*%s*>",
			userStr,
			absBranchURL,
			repo.URI+"@"+event.Branch,
		)
		slack.PostMessage(slack.PostOpts{Msg: msg})
		return
	}

	if payload.Type == GitDeleteEvent {
		msg := fmt.Sprintf("*%s* deleted the branch <%s|*%s*>",
			userStr,
			absBranchURL,
			repo.URI+"@"+event.Branch,
		)
		slack.PostMessage(slack.PostOpts{Msg: msg})
		return
	}

	// See how many commits were pushed.
	commits, err := cl.Repos.ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
		Repo: repo,
		Opt: &sourcegraph.RepoListCommitsOptions{
			Head:         event.Commit,
			Base:         event.Last,
			RefreshCache: true,
			ListOptions:  sourcegraph.ListOptions{PerPage: 1000},
		},
	})
	if err != nil {
		log.Printf("warning: error fetching push commits for post-push hook: %s.", err)
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
		commitURL := router.Rel.URLToRepoCommit(repo.URI, string(c.ID))
		commitMessages = append(commitMessages, fmt.Sprintf("<%s|%s>: %s",
			conf.AppURL(ctx).ResolveReference(commitURL).String(),
			c.ID[:6],
			textutil.ShortCommitMessage(80, c.Message),
		))
	}

	msg := fmt.Sprintf("*%s* pushed *%d %s* to <%s|*%s*>\n%s",
		userStr,
		len(commits.Commits), commitsNoun,
		absBranchURL, repo.URI+"@"+event.Branch,
		strings.Join(commitMessages, "\n"),
	)
	slack.PostMessage(slack.PostOpts{Msg: msg})
}

func buildHook(gitListener *gitHookListener, payload Payload) {
	ctx := gitListener.Ctx
	cl := sourcegraph.NewClientFromContext(ctx)
	repo := payload.Repo
	event := payload.Event
	if event.Type == githttp.PUSH || event.Type == githttp.PUSH_FORCE {
		_, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{
			RepoRev: sourcegraph.RepoRevSpec{RepoSpec: repo, Rev: event.Branch, CommitID: event.Commit},
			Opt:     &sourcegraph.BuildCreateOptions{BuildConfig: sourcegraph.BuildConfig{Import: true, Queue: true}},
		})
		if err != nil {
			log15.Warn("post-push build hook failed to create build", "err", err, "repo", repo.URI, "commit", event.Commit, "branch", event.Branch)
			return
		}
		log15.Debug("post-push build", "repo", repo.URI, "branch", event.Branch, "commit", event.Commit)
	}
}

func getUserDisplayName(cl *sourcegraph.Client, ctx context.Context, actor authpkg.Actor) (string, error) {
	if actor.Login == "" {
		user, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{UID: int32(actor.UID), Domain: actor.Domain})
		if err == nil {
			actor.Login = user.Login
		} else {
			return "", err
		}
	}

	var userStr string
	switch {
	case actor.Login != "":
		userStr = actor.Login
	case actor.UID != 0:
		userStr = fmt.Sprintf("uid %d", actor.UID)
	default:
		userStr = "anonymous user"
	}
	return userStr, nil
}

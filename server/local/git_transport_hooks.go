package local

import (
	"fmt"
	"log"
	"net/url"
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
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/textutil"
)

const GitPushEvent events.EventID = "git.push"

func init() {
	events.Subscribe(GitPushEvent, slackContributionsHook)
	events.Subscribe(GitPushEvent, buildHook)
}

type GitHookPayload struct {
	Ctx    context.Context
	Op     *gitpb.ReceivePackOp
	Events []githttp.Event
}

func slackContributionsHook(payload GitHookPayload) {
	ctx, op, events := payload.Ctx, payload.Op, payload.Events
	userStr, err := getUserDisplayName(ctx)
	if err != nil {
		log.Printf("postPushHook: error getting user: %s", err)
		return
	}

	// Grab a list of all branches so we can identify if one is deleted.
	// TODO(slimsag): find a more canonical way to check if a branch exists?
	repo := op.Repo
	branches, err := svc.Repos(ctx).ListBranches(ctx, &sourcegraph.ReposListBranchesOp{
		Repo: repo,
		Opt:  &sourcegraph.RepoListBranchesOptions{},
	})
	if err != nil {
		log.Printf("warning: error failed to list branches for post-push hook: %s.", err)
	}

	for _, e := range events {
		branchURL, _ := router.Rel.URLToRepoRev(repo.URI, e.Branch)

		// Identify if the branch was deleted.
		haveBranch := false
		for _, b := range branches.Branches {
			if b.Name == e.Branch {
				haveBranch = true
				break
			}
		}
		if !haveBranch {
			msg := fmt.Sprintf("*%s* deleted the branch <%s|*%s*>",
				userStr,
				appURL(ctx, branchURL),
				repo.URI+"@"+e.Branch,
			)
			slack.PostMessage(slack.PostOpts{Msg: msg})
			continue
		}

		// See how many commits were pushed.
		commits, err := svc.Repos(ctx).ListCommits(ctx, &sourcegraph.ReposListCommitsOp{
			Repo: repo,
			Opt: &sourcegraph.RepoListCommitsOptions{
				Head:         e.Commit,
				Base:         e.Last,
				RefreshCache: true,
				ListOptions:  sourcegraph.ListOptions{PerPage: 1000},
			},
		})
		if err != nil {
			log.Printf("warning: error fetching push commits for post-push hook: %s.", err)
			commits = &sourcegraph.CommitList{}
		}

		// Form a nice message if they just pushed a branch with no new commits.
		if len(commits.Commits) == 0 {
			msg := fmt.Sprintf("*%s* created the branch <%s|*%s*>",
				userStr,
				appURL(ctx, branchURL), repo.URI+"@"+e.Branch,
			)
			slack.PostMessage(slack.PostOpts{Msg: msg})
			continue
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
			commitMessages = append(commitMessages, fmt.Sprintf("<%s|%s>: %s", appURL(ctx, commitURL), c.ID[:6], textutil.ShortCommitMessage(80, c.Message)))
		}

		msg := fmt.Sprintf("*%s* pushed *%d %s* to <%s|*%s*>\n%s",
			userStr,
			len(commits.Commits), commitsNoun,
			appURL(ctx, branchURL), repo.URI+"@"+e.Branch,
			strings.Join(commitMessages, "\n"),
		)
		slack.PostMessage(slack.PostOpts{Msg: msg})
	}
}

func getUserDisplayName(ctx context.Context) (string, error) {
	actor := authpkg.ActorFromContext(ctx)
	if actor.Login == "" {
		user, err := svc.Users(ctx).Get(ctx, &sourcegraph.UserSpec{UID: int32(actor.UID), Domain: actor.Domain})
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

func appURL(ctx context.Context, path *url.URL) string {
	return conf.AppURL(ctx).ResolveReference(path).String()
}

func buildHook(payload GitHookPayload) {
	ctx, op, events := payload.Ctx, payload.Op, payload.Events
	for _, e := range events {
		if e.Type == githttp.PUSH || e.Type == githttp.PUSH_FORCE {
			_, err := svc.Builds(ctx).Create(ctx, &sourcegraph.BuildsCreateOp{
				RepoRev: sourcegraph.RepoRevSpec{RepoSpec: op.Repo, Rev: e.Branch, CommitID: e.Commit},
				Opt:     &sourcegraph.BuildCreateOptions{BuildConfig: sourcegraph.BuildConfig{Import: true, Queue: true}},
			})
			if err != nil {
				log15.Warn("post-push build hook failed to create build", "err", err, "repo", op.Repo.URI, "commit", e.Commit, "branch", e.Branch)
				continue
			}
			log15.Debug("post-push build", "repo", op.Repo.URI, "branch", e.Branch, "commit", e.Commit)
		}
	}
}

package local

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/router"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/ext/slack"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/textutil"
)

func init() {
	AddGitPostPushHook("slack.contributions", slackContributionsHook)
	AddGitPostPushHook("build", buildHook)
}

// postPushHooks holds a collection of functions that will be called as git
// post-push hooks. Hooks are identified by their string key.
var postPushHooks = make(map[string]func(context.Context, *gitpb.ReceivePackOp, []githttp.Event))

// errHookExists is returned when the user tries to add a hook with an
// ID that already exists. The `os.IsExist` function can be used to check
// for this error.
var errHookExists = errors.New("ID already exists")

// AddGitPostPushHook adds a new git post-push hook with the given ID. The
// hook receives the context, pack operation information and git events. We
// gaurentee that a hook will never run concurrently with itself. However, it
// does not block the git push and will run concurrently with other hooks. As
// such hooks should not block for a long time, and must handle the repo state
// being different to the advertised events.
func AddGitPostPushHook(ID string, fn func(context.Context, *gitpb.ReceivePackOp, []githttp.Event)) error {
	if _, ok := postPushHooks[ID]; ok {
		return &os.PathError{
			Op:   "AddPostPushHook",
			Path: ID,
			Err:  errHookExists,
		}
	}
	postPushHooks[ID] = linearizePushHook(fn)
	return nil
}

// RemoveGitPostPushHook removes the git hook with the given ID.
func RemoveGitPostPushHook(ID string) { delete(postPushHooks, ID) }

// linearizePushHook wraps a push hook to ensure that the commit hook is never
// run concurrently and is run in FIFO order.
func linearizePushHook(fn func(context.Context, *gitpb.ReceivePackOp, []githttp.Event)) func(context.Context, *gitpb.ReceivePackOp, []githttp.Event) {
	type args struct {
		ctx    context.Context
		op     *gitpb.ReceivePackOp
		events []githttp.Event
	}
	// I don't expect the queue to ever be larger than single digits, it
	// is this high for safety.
	ch := make(chan args, 30)
	go func() {
		a := <-ch
		fn(a.ctx, a.op, a.events)
	}()
	return func(ctx context.Context, op *gitpb.ReceivePackOp, events []githttp.Event) {
		ch <- args{ctx, op, events}
	}
}

func slackContributionsHook(ctx context.Context, op *gitpb.ReceivePackOp, events []githttp.Event) {
	userStr, err := getUserDisplayName(ctx)
	if err != nil {
		log.Printf("pushPushHook: error getting user: %s", err)
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
			go slack.PostMessage(slack.PostOpts{Msg: msg})
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
			go slack.PostMessage(slack.PostOpts{Msg: msg})
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
		go slack.PostMessage(slack.PostOpts{Msg: msg})
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

func buildHook(ctx context.Context, op *gitpb.ReceivePackOp, events []githttp.Event) {
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

package httpapi

import (
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var data struct {
		sourcegraph.RepoList
		sourcegraph.RemoteRepoList
		HasLinkedGitHub bool
		LinkGitHubURL   string
	}

	// true if the user has not yet linked GitHub
	isAuthError := func(err error) bool {
		return grpc.Code(err) == codes.Unauthenticated || grpc.Code(err) == codes.PermissionDenied
	}

	var err error
	var reposOnPage *sourcegraph.RemoteRepoList
	var remoteRepos = &sourcegraph.RemoteRepoList{}
	for page := 1; ; page++ {
		reposOnPage, err = cl.Repos.ListRemote(ctx, &sourcegraph.ReposListRemoteOptions{
			ListOptions: sourcegraph.ListOptions{PerPage: 100, Page: int32(page)},
		})
		if err != nil {
			break
		}

		if len(reposOnPage.RemoteRepos) == 0 {
			break
		}
		remoteRepos.RemoteRepos = append(remoteRepos.RemoteRepos, reposOnPage.RemoteRepos...)
	}

	if err != nil {
		if !isAuthError(err) {
			return err
		}
	} else {
		data.HasLinkedGitHub = true
		data.RemoteRepos = remoteRepos.RemoteRepos
	}
	// TODO(john): provide static routes to client in context.
	data.LinkGitHubURL = router.Rel.URLTo(router.GitHubOAuth2Initiate).String()

	// Only show hosted repos on the dashboard to users that have write access
	// on this server.
	if auth.ActorFromContext(ctx).HasWriteAccess() {
		hostedRepos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
			Sort:        "pushed",
			Direction:   "desc",
			ListOptions: sourcegraph.ListOptions{PerPage: 100},
		})
		if err != nil {
			return err
		} else {
			data.Repos = hostedRepos.Repos
		}
	} else {
		data.Repos = []*sourcegraph.Repo{}
	}

	return writeJSON(w, &data)
}

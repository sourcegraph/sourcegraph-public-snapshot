package app

import (
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	currentUser := handlerutil.UserFromRequest(r)

	var data struct {
		Repos       *sourcegraph.RepoList
		RemoteRepos *sourcegraph.RemoteRepoList
		LinkGitHub  bool
		Teammates   *sourcegraph.Teammates

		// This flag is set if the user has returned to the dashboard after
		// being redirected from GitHub OAuth2 login page.
		GitHubOnboarding bool

		tmpl.Common
	}

	data.GitHubOnboarding = r.URL.Query().Get("github-onboarding") == "true"

	isAuthError := func(err error) bool {
		return grpc.Code(err) == codes.Unauthenticated || grpc.Code(err) == codes.PermissionDenied
	}

	var err error
	var reposOnPage *sourcegraph.RemoteRepoList
	data.RemoteRepos = &sourcegraph.RemoteRepoList{}
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
		data.RemoteRepos.RemoteRepos = append(data.RemoteRepos.RemoteRepos, reposOnPage.RemoteRepos...)
	}

	if isAuthError(err) {
		data.LinkGitHub = true
	} else if err != nil {
		return err
	}

	if currentUser != nil {
		data.Teammates, err = cl.Users.ListTeammates(ctx, currentUser)
		if isAuthError(err) {
			data.LinkGitHub = true
		} else if err != nil {
			return err
		}
	}

	// Only show hosted repos on the dashboard to users that have write access
	// on this server.
	if auth.ActorFromContext(ctx).HasWriteAccess() {
		data.Repos, err = cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
			Sort:        "pushed",
			Direction:   "desc",
			ListOptions: sourcegraph.ListOptions{PerPage: 100},
		})
		if err != nil {
			return err
		}
	} else {
		data.Repos = &sourcegraph.RepoList{
			Repos: []*sourcegraph.Repo{},
		}
	}

	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, &data)
}

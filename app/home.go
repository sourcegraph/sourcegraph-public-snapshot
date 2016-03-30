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

	// TODO(sqs): add pagination
	listOpt := sourcegraph.ListOptions{PerPage: 250}

	isAuthError := func(err error) bool {
		return grpc.Code(err) == codes.Unauthenticated || grpc.Code(err) == codes.PermissionDenied
	}

	var err error
	data.RemoteRepos, err = cl.Repos.ListRemote(ctx, &sourcegraph.ReposListRemoteOptions{ListOptions: listOpt})
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
			ListOptions: listOpt,
		})
		if err != nil {
			return err
		}
	} else {
		// Show Go starter repo if it is available.
		repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "sample/golang/hello"})
		if err == nil {
			data.Repos = &sourcegraph.RepoList{
				Repos: []*sourcegraph.Repo{repo},
			}
		}
	}

	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, &data)
}

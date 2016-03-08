package app

import (
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	appauthutil "src.sourcegraph.com/sourcegraph/app/internal/authutil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	currentUser := handlerutil.UserFromRequest(r)

	if currentUser == nil && authutil.ActiveFlags.HasUserAccounts() {
		return appauthutil.RedirectToLogIn(w, r)
	}

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
	listOpt := sourcegraph.ListOptions{PerPage: 100}

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

	data.Repos, err = cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: listOpt,
	})
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, &data)
}

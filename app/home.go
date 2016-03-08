package app

import (
	"net/http"

	"sourcegraph.com/sqs/pbtypes"
	appauthutil "src.sourcegraph.com/sourcegraph/app/internal/authutil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

type dashboardData struct {
	Repos      []*sourcegraph.Repo
	LinkGitHub bool
	MirrorData *sourcegraph.UserMirrorData
	Teammates  *sourcegraph.Teammates

	// This flag is set if the user has returned to the dashboard after
	// being redirected from GitHub OAuth2 login page.
	GitHubOnboarding bool

	tmpl.Common
}

type marshalledDashboard struct {
	JSON string
}

func execDashboardTmpl(w http.ResponseWriter, r *http.Request, d *dashboardData) error {
	q := r.URL.Query()
	if q.Get("github-onboarding") == "true" {
		d.GitHubOnboarding = true
	}
	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, d)
}

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	currentUser := handlerutil.UserFromRequest(r)

	if currentUser == nil && authutil.ActiveFlags.HasUserAccounts() {
		return appauthutil.RedirectToLogIn(w, r)
	}

	var mirrorData *sourcegraph.UserMirrorData
	var teammates *sourcegraph.Teammates

	if authutil.ActiveFlags.HasUserAccounts() {
		var err error
		mirrorData, err = cl.MirrorRepos.GetUserData(ctx, &pbtypes.Void{})
		if err != nil {
			return err
		}

		switch mirrorData.State {
		case sourcegraph.UserMirrorsState_NoToken, sourcegraph.UserMirrorsState_InvalidToken:
			return execDashboardTmpl(w, r, &dashboardData{
				MirrorData: mirrorData,
				LinkGitHub: true,
			})
		}

		teammates, err = cl.Users.ListTeammates(ctx, currentUser)
		if err != nil {
			return err
		}
	}

	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: sourcegraph.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}

	return execDashboardTmpl(w, r, &dashboardData{
		Repos:      repos.Repos,
		MirrorData: mirrorData,
		Teammates:  teammates,
	})
}

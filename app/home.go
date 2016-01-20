package app

import (
	"net/http"
	"os"

	"golang.org/x/net/context"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	currentUser := handlerutil.UserFromRequest(r)

	allowPrivateMirrors := authutil.ActiveFlags.HasPrivateMirrors()
	var gd *sourcegraph.GitHubRepoData
	var teammates []*userInfo
	if allowPrivateMirrors && currentUser != nil {
		var err error
		gd, err = cl.Repos.GetPrivateGitHubRepos(ctx, &sourcegraph.GitHubRepoRequest{})
		if err != nil {
			return err
		}

		teammates = getTeammates(ctx, cl)
	}

	var listOpts sourcegraph.ListOptions
	if err := schemautil.Decode(&listOpts, r.URL.Query()); err != nil {
		return err
	}

	if listOpts.PerPage == 0 {
		listOpts.PerPage = 50
	}

	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: listOpts,
	})
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, &struct {
		Repos  []*sourcegraph.Repo
		SGPath string
		Users  []*userInfo
		IsLDAP bool

		AllowPrivateMirrors bool
		GitHub              *sourcegraph.GitHubRepoData
		Teammates           []*userInfo

		tmpl.Common
	}{
		Repos:  repos.Repos,
		SGPath: os.Getenv("SGPATH"),
		Users:  getUsersAndInvites(ctx, cl),
		IsLDAP: authutil.ActiveFlags.IsLDAP(),

		AllowPrivateMirrors: allowPrivateMirrors,
		GitHub:              gd,
		Teammates:           teammates,
	})
}

type userInfo struct {
	Identifier string
	AvatarURL  string
	Write      bool
	Admin      bool
	Invite     bool
}

func getUsersAndInvites(ctx context.Context, cl *sourcegraph.Client) []*userInfo {
	var users []*userInfo
	ctxActor := auth.ActorFromContext(ctx)
	if !ctxActor.HasAdminAccess() { // current user is not an admin of the instance
		return users
	}

	// Fetch pending invites.
	inviteList, err := cl.Accounts.ListInvites(ctx, &pbtypes.Void{})
	if err == nil {
		for _, invite := range inviteList.Invites {
			users = append(users, &userInfo{
				Identifier: invite.Email,
				Write:      invite.Write,
				Admin:      invite.Admin,
				Invite:     true,
			})
		}
	}

	// Fetch registered users.
	userList, err := cl.Users.List(ctx, &sourcegraph.UsersListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: 10000,
		},
	})
	if err == nil {
		for _, user := range userList.Users {
			users = append(users, &userInfo{
				Identifier: user.Login,
				AvatarURL:  user.AvatarURL,
				Write:      user.Write,
				Admin:      user.Admin,
			})
		}
	}
	return users
}

func getTeammates(ctx context.Context, cl *sourcegraph.Client) []*userInfo {
	var teammates []*userInfo

	return teammates
}

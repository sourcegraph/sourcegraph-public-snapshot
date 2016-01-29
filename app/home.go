package app

import (
	"net/http"

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

type dashboardData struct {
	Repos       []*sourcegraph.Repo
	Users       []*userInfo
	IsLDAP      bool
	MirrorsNext bool
	LinkGitHub  bool
	OnWaitlist  bool
	MirrorData  *sourcegraph.UserMirrorData
	Teammates   *sourcegraph.Teammates

	tmpl.Common
}

func execDashboardTmpl(w http.ResponseWriter, r *http.Request, d *dashboardData) error {
	d.IsLDAP = authutil.ActiveFlags.IsLDAP()
	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, d)
}

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	currentUser := handlerutil.UserFromRequest(r)
	var users []*userInfo

	mirrorsNext := currentUser != nil && authutil.ActiveFlags.MirrorsNext
	var mirrorData *sourcegraph.UserMirrorData
	var teammates *sourcegraph.Teammates
	if mirrorsNext {
		var err error
		mirrorData, err = cl.MirrorRepos.GetUserData(ctx, &pbtypes.Void{})
		if err != nil {
			return err
		}

		switch mirrorData.State {
		case sourcegraph.UserMirrorsState_NoToken, sourcegraph.UserMirrorsState_InvalidToken:
			return execDashboardTmpl(w, r, &dashboardData{
				MirrorsNext: mirrorsNext,
				MirrorData:  mirrorData,
				LinkGitHub:  true,
			})
		case sourcegraph.UserMirrorsState_OnWaitlist:
			return execDashboardTmpl(w, r, &dashboardData{
				MirrorsNext: mirrorsNext,
				MirrorData:  mirrorData,
				OnWaitlist:  true,
			})
		case sourcegraph.UserMirrorsState_NotAllowed:
			mirrorsNext = false
		}
	}
	if mirrorsNext {
		var err error
		teammates, err = cl.Users.ListTeammates(ctx, currentUser)
		if err != nil {
			return err
		}
	} else {
		users = getUsers(ctx, cl)
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

	return execDashboardTmpl(w, r, &dashboardData{
		Repos:       repos.Repos,
		Users:       users,
		MirrorsNext: mirrorsNext,
		MirrorData:  mirrorData,
		Teammates:   teammates,
	})
}

type userInfo struct {
	UID       int32
	Name      string
	Login     string
	AvatarURL string
	Write     bool
	Admin     bool
	Invite    bool
}

func getUsers(ctx context.Context, cl *sourcegraph.Client) []*userInfo {
	var users []*userInfo
	ctxActor := auth.ActorFromContext(ctx)
	if !ctxActor.HasAdminAccess() {
		return users
	}

	// Fetch pending invites.
	inviteList, err := cl.Accounts.ListInvites(ctx, &pbtypes.Void{})
	if err == nil {
		for _, invite := range inviteList.Invites {
			users = append(users, &userInfo{
				Name:   invite.Email,
				Write:  invite.Write,
				Admin:  invite.Admin,
				Invite: true,
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
				UID:       user.UID,
				Name:      user.Name,
				Login:     user.Login,
				AvatarURL: user.AvatarURL,
				Write:     user.Write,
				Admin:     user.Admin,
			})
		}
	}
	return users
}

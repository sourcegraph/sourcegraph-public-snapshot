package app

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
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

	mirrorsNext := currentUser != nil && authutil.ActiveFlags.MirrorsNext
	var mirrorData *sourcegraph.UserMirrorData
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
	users := getUsers(ctx, cl, mirrorsNext)

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
	})
}

type userInfo struct {
	UID        int32
	Identifier string
	AvatarURL  string
	Write      bool
	Admin      bool
	Invite     bool
}

func getUsers(ctx context.Context, cl *sourcegraph.Client, allowMirrorsNext bool) []*userInfo {
	var users []*userInfo
	ctxActor := auth.ActorFromContext(ctx)
	if !allowMirrorsNext && !ctxActor.HasAdminAccess() {
		return users
	}
	users = fetchUsersAndInvites(ctx, cl)
	if !allowMirrorsNext {
		return users
	}

	teammates := fetchTeammates(ctx, cl, int32(ctxActor.UID), users)
	return teammates
}

func fetchUsersAndInvites(ctx context.Context, cl *sourcegraph.Client) []*userInfo {
	var users []*userInfo

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
				UID:        user.UID,
				Identifier: getIdentifier(user.Name, user.Login),
				AvatarURL:  user.AvatarURL,
				Write:      user.Write,
				Admin:      user.Admin,
			})
		}
	}
	return users
}

func fetchTeammates(ctx context.Context, cl *sourcegraph.Client, currentUID int32, users []*userInfo) []*userInfo {
	var teammates []*userInfo

	// Fetch the currently authenticated user's stored access token (if any).
	extToken, err := cl.Auth.GetExternalToken(ctx, &sourcegraph.ExternalTokenRequest{
		Host: githubcli.Config.Host(),
	})
	if err != nil {
		return teammates
	}

	client := githubutil.Default.AuthedClient(extToken.Token)

	extAccountsStore := auth.ExtAccountsStore{}
	ghOrgs, _, err := client.Organizations.List("", nil)
	if err != nil {
		return teammates
	}

	linkedAccounts := make(map[int32]bool)

	for _, org := range ghOrgs {
		acct, err := extAccountsStore.GetAll(ctx, githubcli.Config.Host(), *org.Login)
		if err != nil {
			continue
		}

		var foundSelf bool
		for _, member := range acct.Users {
			if currentUID == member {
				foundSelf = true
			}
			linkedAccounts[member] = true
		}

		// Add this user to the organization's cached member list.
		if !foundSelf {
			if err := extAccountsStore.Append(ctx, githubcli.Config.Host(), *org.Login, currentUID); err != nil {
				log15.Debug("Could not add user to org's member list", "uid", currentUID, "org", *org.Login, "error", err)
			}
		}
	}

	for _, user := range users {
		if _, ok := linkedAccounts[user.UID]; ok {
			teammates = append(teammates, user)
		}
	}

	return teammates
}

func getIdentifier(name, login string) string {
	if name == "" {
		return login
	} else {
		return fmt.Sprintf("%s (%s)", name, login)
	}
}

package local

import (
	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

func (s *users) ListTeammates(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.Teammates, error) {
	if user.UID == 0 {
		return nil, grpc.Errorf(codes.FailedPrecondition, "no uid specified")
	}

	// This call will also confirm that the request actor has access to this user's teammate info.
	extToken, err := svc.Auth(ctx).GetExternalToken(ctx, &sourcegraph.ExternalTokenRequest{UID: user.UID})
	if grpc.Code(err) == codes.NotFound {
		return &sourcegraph.Teammates{}, nil
	} else if err != nil {
		return nil, err
	}

	extTokenStore := store.ExternalAuthTokensFromContext(ctx)
	usersStore := store.UsersFromContext(ctx)

	client := githubutil.Default.AuthedClient(extToken.Token)
	githubCtx := github.NewContextWithClient(ctx, client)

	ghOrgsStore := github.Orgs{}
	ghOrgs, err := ghOrgsStore.List(githubCtx, sourcegraph.UserSpec{}, &sourcegraph.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	// Record the user's GitHub orgs
	var ghOrgNames []string
	for _, org := range ghOrgs {
		ghOrgNames = append(ghOrgNames, org.Login)
	}
	if err := store.WaitlistFromContext(ctx).UpdateUserOrgs(elevatedActor(ctx), user.UID, ghOrgNames); err != nil {
		log15.Warn("Could not record user's GitHub orgs", "uid", user.UID, "error", err)
	}

	usersByOrg := make(map[string]*sourcegraph.RemoteUserList)
	githubMembers := make(map[int32]struct{})
	githubUIDs := make([]int, 0)
	for _, org := range ghOrgs {
		members, err := ghOrgsStore.ListMembers(githubCtx, sourcegraph.OrgSpec{Org: org.Login}, &sourcegraph.OrgListMembersOptions{
			ListOptions: sourcegraph.ListOptions{PerPage: 1000},
		})
		if err != nil {
			log15.Warn("Could not list members for GitHub org", "org", org.Login, "error", err)
			continue
		}
		usersByOrg[org.Login] = &sourcegraph.RemoteUserList{
			Users: make([]*sourcegraph.RemoteUser, len(members)),
		}

		for i, m := range members {
			if _, ok := githubMembers[m.UID]; !ok {
				githubMembers[m.UID] = struct{}{}
				githubUIDs = append(githubUIDs, int(m.UID))
			}

			usersByOrg[org.Login].Users[i] = &sourcegraph.RemoteUser{
				RemoteAccount: m,
			}
			// Fetch the primary email of the GitHub user.
			ghuser, _, err := client.Users.Get(m.Login)
			if err != nil {
				log15.Warn("Could not fetch github user", "login", m.Login, "error", err)
				continue
			}
			if ghuser.Name != nil {
				usersByOrg[org.Login].Users[i].RemoteAccount.Name = *ghuser.Name
			}
			if ghuser.Email != nil {
				usersByOrg[org.Login].Users[i].Email = *ghuser.Email
			}
		}
	}
	linkedUserTokens, err := extTokenStore.ListExternalUsers(elevatedActor(ctx), githubUIDs, githubcli.Config.Host(), githubClientID)
	if err != nil {
		return nil, err
	}

	// uidMap maps a github UID to the list of UIDs of Sourcegraph user
	// accounts that are linked to that GitHub account.
	uidMap := make(map[int32][]int32)
	sgUIDs := make([]int32, 0)
	for _, tok := range linkedUserTokens {
		ghID := int32(tok.ExtUID)
		sgID := int32(tok.User)
		if _, ok := uidMap[ghID]; !ok {
			uidMap[ghID] = make([]int32, 0)
		}
		uidMap[ghID] = append(uidMap[ghID], sgID)
		sgUIDs = append(sgUIDs, sgID)
	}

	sgUsers, err := usersStore.List(elevatedActor(ctx), &sourcegraph.UsersListOptions{UIDs: sgUIDs})
	if err != nil {
		return nil, err
	}

	sgUserMap := make(map[int32]*sourcegraph.User)
	for _, u := range sgUsers {
		sgUserMap[u.UID] = u
	}

	// Fetch pending invites.
	invitesMap := make(map[string]struct{})
	inviteList, err := svc.Accounts(ctx).ListInvites(elevatedActor(ctx), &pbtypes.Void{})
	if err == nil {
		for _, invite := range inviteList.Invites {
			invitesMap[invite.Email] = struct{}{}
		}
	}

	for orgName := range usersByOrg {
		for i := range usersByOrg[orgName].Users {
			ghUID := usersByOrg[orgName].Users[i].RemoteAccount.UID
			if sgUIDs, ok := uidMap[ghUID]; ok {
				for _, id := range sgUIDs {
					// TODO: make a new RemoteUser for every Sourcegraph user
					// linked to the same GitHub account
					if sgUser, ok := sgUserMap[id]; ok {
						usersByOrg[orgName].Users[i].LocalAccount = sgUser
					}
				}
			}

			// Check if there is a pending invite for this user.
			ghEmail := usersByOrg[orgName].Users[i].Email
			if ghEmail != "" {
				if _, ok := invitesMap[ghEmail]; ok {
					usersByOrg[orgName].Users[i].IsInvited = true
				}
			}
		}
	}

	return &sourcegraph.Teammates{UsersByOrg: usersByOrg}, nil
}

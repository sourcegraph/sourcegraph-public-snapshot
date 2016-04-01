package local

import (
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/ext/github"
	"sourcegraph.com/sourcegraph/sourcegraph/ext/github/githubcli"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/util/githubutil"
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
	githubCtx := github.NewContextWithClient(ctx, client, true)

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

	var userList []*sourcegraph.RemoteUser
	for _, org := range ghOrgs {
		members, err := ghOrgsStore.ListMembers(githubCtx, sourcegraph.OrgSpec{Org: org.Login}, &sourcegraph.OrgListMembersOptions{
			ListOptions: sourcegraph.ListOptions{PerPage: 1000},
		})
		if err != nil {
			log15.Warn("Could not list members for GitHub org", "org", org.Login, "error", err)
			continue
		}

		var wg sync.WaitGroup
		for i := range members {
			currentOrgLogin := org.Login
			currentUser := &sourcegraph.RemoteUser{
				RemoteAccount: members[i],
				Organization:  currentOrgLogin,
			}
			userList = append(userList, currentUser)

			wg.Add(1)
			go func() {
				defer wg.Done()
				// Fetch the primary email of the GitHub user.
				// Use a client local to this goroutine since it is not thread-safe.
				client := githubutil.Default.AuthedClient(extToken.Token)
				ghuser, _, err := client.Users.Get(currentUser.RemoteAccount.Login)
				if err != nil {
					log15.Warn("Could not fetch github user", "login", currentUser.RemoteAccount.Login, "error", err)
					return
				}
				if ghuser.Name != nil {
					currentUser.RemoteAccount.Name = *ghuser.Name
				}
				if ghuser.Email != nil {
					currentUser.Email = *ghuser.Email
				}
			}()
		}
		wg.Wait()
	}

	if len(userList) == 0 {
		currentUser, _ := usersStore.Get(ctx, *user)
		userList = append(userList, &sourcegraph.RemoteUser{
			RemoteAccount: currentUser,
			Organization:  currentUser.Login,
		})
	}

	var githubUIDs []int
	for _, user := range userList {
		githubUIDs = append(githubUIDs, int(user.RemoteAccount.UID))
	}
	linkedUserTokens, err := extTokenStore.ListExternalUsers(elevatedActor(ctx), githubUIDs, githubcli.Config.Host(), githubClientID)
	if err != nil {
		return nil, err
	}

	// uidMap maps a github UID to the list of UIDs of Sourcegraph user
	// accounts that are linked to that GitHub account.
	uidMap := make(map[int32][]int32)
	var sgUIDs []int32
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

	for _, user := range userList {
		ghUID := user.RemoteAccount.UID
		if sgUIDs, ok := uidMap[ghUID]; ok {
			for _, id := range sgUIDs {
				// TODO: make a new RemoteUser for every Sourcegraph user
				// linked to the same GitHub account
				if sgUser, ok := sgUserMap[id]; ok {
					user.LocalAccount = sgUser
				}
			}
		}
	}

	return &sourcegraph.Teammates{Users: userList, Organizations: ghOrgNames}, nil
}

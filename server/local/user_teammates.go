package local

import (
	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
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

	extTokenStore := store.ExternalAuthTokensFromContextOrNil(ctx)
	if extTokenStore == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "ext auth tokens store not implemented")
	}
	usersStore := store.UsersFromContextOrNil(ctx)
	if usersStore == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "users store not implemented")
	}

	client := githubutil.Default.AuthedClient(extToken.Token)
	githubCtx := github.NewContextWithClient(ctx, client)

	ghOrgsStore := github.Orgs{}
	ghOrgs, err := ghOrgsStore.List(githubCtx, sourcegraph.UserSpec{}, &sourcegraph.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	githubMembers := make(map[int32]*sourcegraph.User)
	githubUIDs := make([]int, 0)
	for _, org := range ghOrgs {
		members, err := ghOrgsStore.ListMembers(githubCtx, sourcegraph.OrgSpec{Org: org.Login}, &sourcegraph.OrgListMembersOptions{
			ListOptions: sourcegraph.ListOptions{PerPage: 1000},
		})
		if err != nil {
			log15.Warn("Could not list members for GitHub org", "org", org.Login, "error", err)
			continue
		}

		for _, m := range members {
			if _, ok := githubMembers[m.UID]; !ok {
				githubMembers[m.UID] = m
				githubUIDs = append(githubUIDs, int(m.UID))
			}
		}
	}
	linkedUserTokens, err := extTokenStore.ListExternalUsers(elevatedActor(ctx), githubUIDs, githubcli.Config.Host(), githubClientID)
	if err != nil {
		return nil, err
	}
	sgUIDs := make([]int32, 0)
	for _, tok := range linkedUserTokens {
		sgUIDs = append(sgUIDs, int32(tok.User))
	}

	sgUsers, err := usersStore.List(elevatedActor(ctx), &sourcegraph.UsersListOptions{UIDs: sgUIDs})
	if err != nil {
		return nil, err
	}

	extUsers := make([]*sourcegraph.User, 0)
	sgUIDMap := make(map[int32]struct{})
	for _, uid := range sgUIDs {
		sgUIDMap[uid] = struct{}{}
	}
	for ghUID, ghUser := range githubMembers {
		if _, ok := sgUIDMap[ghUID]; !ok {
			extUsers = append(extUsers, ghUser)
		}
	}

	veryShortCache(ctx)
	return &sourcegraph.Teammates{
		LinkedUsers:   sgUsers,
		ExternalUsers: extUsers,
	}, nil
}

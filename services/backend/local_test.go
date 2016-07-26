package backend

import (
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/mock"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	githubmock "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/mocks"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

// testContext creates a new context.Context for use by tests that has
// all mockstores instantiated.
func testContext() (context.Context, *mocks) {
	var m mocks
	ctx := context.Background()
	ctx = store.WithStores(ctx, m.stores.Stores())
	ctx = svc.WithServices(ctx, m.servers.servers())
	ctx = github.WithRepos(ctx, &m.githubRepos)
	ctx = conf.WithURL(ctx, &url.URL{Scheme: "http", Host: "example.com"})
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: 1, Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)
	return ctx, &m
}

type mocks struct {
	stores      mockstore.Stores
	servers     mockServers
	githubRepos githubmock.GitHubRepoGetter
}

type mockServers struct {
	Accounts     mock.AccountsServer
	Auth         mock.AuthServer
	Builds       mock.BuildsServer
	Defs         mock.DefsServer
	MirrorRepos  mock.MirrorReposServer
	RepoStatuses mock.RepoStatusesServer
	RepoTree     mock.RepoTreeServer
	Repos        mock.ReposServer
	Search       mock.SearchServer
	Users        mock.UsersServer
}

func (s *mockServers) servers() svc.Services {
	return svc.Services{
		Accounts:     &s.Accounts,
		Auth:         &s.Auth,
		Builds:       &s.Builds,
		Defs:         &s.Defs,
		MirrorRepos:  &s.MirrorRepos,
		RepoStatuses: &s.RepoStatuses,
		RepoTree:     &s.RepoTree,
		Repos:        &s.Repos,
		Search:       &s.Search,
		Users:        &s.Users,
	}
}

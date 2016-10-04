package backend

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/mock"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	githubmock "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/mocks"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

// testContext creates a new context.Context for use by tests that has
// all mockstores instantiated.
func testContext() (context.Context, *mocks) {
	var m mocks
	ctx := context.Background()
	localstore.Mocks = localstore.MockStores{}
	localstore.Graph = &localstore.Mocks.Graph
	ctx = svc.WithServices(ctx, m.servers.servers())
	ctx = github.WithRepos(ctx, &m.githubRepos)
	ctx = authpkg.WithActor(ctx, &authpkg.Actor{UID: "1", Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")
	return ctx, &m
}

type mocks struct {
	servers     mockServers
	githubRepos githubmock.GitHubRepoGetter
}

type mockServers struct {
	Async        mock.AsyncServer
	Defs         mock.DefsServer
	MirrorRepos  mock.MirrorReposServer
	RepoStatuses mock.RepoStatusesServer
	RepoTree     mock.RepoTreeServer
	Repos        mock.ReposServer
	Search       mock.SearchServer
}

func (s *mockServers) servers() svc.Services {
	return svc.Services{
		Async:        &s.Async,
		Defs:         &s.Defs,
		MirrorRepos:  &s.MirrorRepos,
		RepoStatuses: &s.RepoStatuses,
		RepoTree:     &s.RepoTree,
		Repos:        &s.Repos,
		Search:       &s.Search,
	}
}

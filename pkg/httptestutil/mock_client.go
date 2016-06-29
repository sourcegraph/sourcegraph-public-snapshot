package httptestutil

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/mock"
)

type MockClients struct {
	// Ctx is the base context passed to test HTTP handlers. Test code
	// can modify this field to inject context into test HTTP
	// handlers.
	Ctx context.Context

	// TODO(sqs): move this to ./api/sourcegraph
	Annotations  mock.AnnotationsClient
	Accounts     mock.AccountsClient
	Async        mock.AsyncClient
	Auth         mock.AuthClient
	Builds       mock.BuildsClient
	Defs         mock.DefsClient
	Deltas       mock.DeltasClient
	Meta         mock.MetaClient
	MirrorRepos  mock.MirrorReposClient
	Orgs         mock.OrgsClient
	People       mock.PeopleClient
	RepoStatuses mock.RepoStatusesClient
	RepoTree     mock.RepoTreeClient
	Repos        mock.ReposClient
	Search       mock.SearchClient
	Users        mock.UsersClient
}

func (c *MockClients) Client() *sourcegraph.Client {
	return &sourcegraph.Client{
		Annotations:  &c.Annotations,
		Accounts:     &c.Accounts,
		Async:        &c.Async,
		Auth:         &c.Auth,
		Builds:       &c.Builds,
		Defs:         &c.Defs,
		Deltas:       &c.Deltas,
		Meta:         &c.Meta,
		MirrorRepos:  &c.MirrorRepos,
		Orgs:         &c.Orgs,
		People:       &c.People,
		RepoStatuses: &c.RepoStatuses,
		RepoTree:     &c.RepoTree,
		Repos:        &c.Repos,
		Search:       &c.Search,
		Users:        &c.Users,
	}
}

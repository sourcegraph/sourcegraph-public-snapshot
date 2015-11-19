package httptestutil

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
)

type MockClients struct {
	// Ctx is the base context passed to test HTTP handlers. Test code
	// can modify this field to inject context into test HTTP
	// handlers.
	Ctx context.Context

	// TODO(sqs): move this to go-sourcegraph
	Accounts          mock.AccountsClient
	Auth              mock.AuthClient
	Builds            mock.BuildsClient
	Defs              mock.DefsClient
	Deltas            mock.DeltasClient
	Markdown          mock.MarkdownClient
	Meta              mock.MetaClient
	MirrorRepos       mock.MirrorReposClient
	Orgs              mock.OrgsClient
	People            mock.PeopleClient
	RegisteredClients mock.RegisteredClientsClient
	RepoBadges        mock.RepoBadgesClient
	RepoStatuses      mock.RepoStatusesClient
	RepoTree          mock.RepoTreeClient
	Repos             mock.ReposClient
	Search            mock.SearchClient
	Units             mock.UnitsClient
	Users             mock.UsersClient
}

func (c *MockClients) Client() *sourcegraph.Client {
	return &sourcegraph.Client{
		Accounts:          &c.Accounts,
		Auth:              &c.Auth,
		Builds:            &c.Builds,
		Defs:              &c.Defs,
		Deltas:            &c.Deltas,
		Markdown:          &c.Markdown,
		Meta:              &c.Meta,
		MirrorRepos:       &c.MirrorRepos,
		Orgs:              &c.Orgs,
		People:            &c.People,
		RegisteredClients: &c.RegisteredClients,
		RepoBadges:        &c.RepoBadges,
		RepoStatuses:      &c.RepoStatuses,
		RepoTree:          &c.RepoTree,
		Repos:             &c.Repos,
		Search:            &c.Search,
		Units:             &c.Units,
		Users:             &c.Users,
	}
}

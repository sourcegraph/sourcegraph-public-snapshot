package httptestutil

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/mock"
)

type MockClients struct {
	// Ctx is the base context passed to test HTTP handlers. Test code
	// can modify this field to inject context into test HTTP
	// handlers.
	Ctx context.Context

	Annotations  mock.AnnotationsClient
	Async        mock.AsyncClient
	Defs         mock.DefsClient
	Meta         mock.MetaClient
	MirrorRepos  mock.MirrorReposClient
	RepoStatuses mock.RepoStatusesClient
	RepoTree     mock.RepoTreeClient
	Repos        mock.ReposClient
	Search       mock.SearchClient
}

func (c *MockClients) Client() *sourcegraph.Client {
	return &sourcegraph.Client{
		Annotations:  &c.Annotations,
		Async:        &c.Async,
		Defs:         &c.Defs,
		Meta:         &c.Meta,
		MirrorRepos:  &c.MirrorRepos,
		RepoStatuses: &c.RepoStatuses,
		RepoTree:     &c.RepoTree,
		Repos:        &c.Repos,
		Search:       &c.Search,
	}
}

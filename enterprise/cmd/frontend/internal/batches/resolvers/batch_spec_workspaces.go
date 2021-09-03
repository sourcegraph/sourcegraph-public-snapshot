package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type batchSpecWorkspacesResolver struct {
	store            *store.Store
	allowUnsupported bool
	allowIgnored     bool
	unsupported      map[*types.Repo]struct{}
	ignored          map[*types.Repo]struct{}
	workspaces       []*service.RepoWorkspace
}

var _ graphqlbackend.BatchSpecWorkspacesResolver = &batchSpecWorkspacesResolver{}

func (r *batchSpecWorkspacesResolver) AllowIgnored() bool {
	return r.allowIgnored
}

func (r *batchSpecWorkspacesResolver) AllowUnsupported() bool {
	return r.allowUnsupported
}

func (r *batchSpecWorkspacesResolver) Workspaces() []graphqlbackend.BatchSpecWorkspaceResolver {
	resolvers := make([]graphqlbackend.BatchSpecWorkspaceResolver, 0, len(r.workspaces))
	for _, node := range r.workspaces {
		node := node
		resolvers = append(resolvers, &batchSpecWorkspaceResolver{node: node, store: r.store})
	}
	return resolvers
}

func (r *batchSpecWorkspacesResolver) Unsupported() []*graphqlbackend.RepositoryResolver {
	resolvers := make([]*graphqlbackend.RepositoryResolver, 0, len(r.unsupported))
	for repo := range r.unsupported {
		resolvers = append(resolvers, graphqlbackend.NewRepositoryResolver(r.store.DB(), repo))
	}
	return resolvers
}

func (r *batchSpecWorkspacesResolver) Ignored() []*graphqlbackend.RepositoryResolver {
	resolvers := make([]*graphqlbackend.RepositoryResolver, 0, len(r.ignored))
	for repo := range r.ignored {
		resolvers = append(resolvers, graphqlbackend.NewRepositoryResolver(r.store.DB(), repo))
	}
	return resolvers
}

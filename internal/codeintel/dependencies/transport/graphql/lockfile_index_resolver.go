package graphql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
)

type LockfileIndexResolver struct {
	lockfile shared.LockfileIndex
	repo     *graphqlbackend.RepositoryResolver
	commit   *graphqlbackend.GitCommitResolver
}

func NewLockfileIndexResolver(lockfile shared.LockfileIndex, repo *graphqlbackend.RepositoryResolver, commit *graphqlbackend.GitCommitResolver) graphqlbackend.LockfileIndexResolver {
	return &LockfileIndexResolver{lockfile: lockfile, repo: repo, commit: commit}
}

func (e *LockfileIndexResolver) ID() graphql.ID {
	return relay.MarshalID("LockfileIndex", (int64(e.lockfile.ID)))
}

func (e *LockfileIndexResolver) Lockfile() string {
	return e.lockfile.Lockfile
}

func (e *LockfileIndexResolver) Repository() *graphqlbackend.RepositoryResolver {
	return e.repo
}

func (e *LockfileIndexResolver) Commit() *graphqlbackend.GitCommitResolver {
	return e.commit
}

func (e *LockfileIndexResolver) Fidelity() string {
	return e.lockfile.Fidelity.ToGraphQL()
}

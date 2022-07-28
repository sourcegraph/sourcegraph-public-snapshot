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

const lockfileIndexIDKind = "LockfileIndex"

func NewLockfileIndexResolver(lockfile shared.LockfileIndex, repo *graphqlbackend.RepositoryResolver, commit *graphqlbackend.GitCommitResolver) graphqlbackend.LockfileIndexResolver {
	return &LockfileIndexResolver{lockfile: lockfile, repo: repo, commit: commit}
}

func (e *LockfileIndexResolver) ID() graphql.ID {
	return relay.MarshalID(lockfileIndexIDKind, (int64(e.lockfile.ID)))
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

func (e *LockfileIndexResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: e.lockfile.CreatedAt}
}

func (e *LockfileIndexResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: e.lockfile.UpdatedAt}
}

func unmarshalLockfileIndexID(id graphql.ID) (lockfileIndexID int, err error) {
	err = relay.UnmarshalSpec(id, &lockfileIndexID)
	return
}

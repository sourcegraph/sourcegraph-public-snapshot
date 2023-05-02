package gitresolvers

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type commitResolver struct {
	repo resolverstubs.RepositoryResolver
	oid  resolverstubs.GitObjectID
	rev  string
}

func NewGitCommitResolver(repo resolverstubs.RepositoryResolver, commitID api.CommitID, inputRev string) resolverstubs.GitCommitResolver {
	rev := string(commitID)
	if inputRev != "" {
		rev = inputRev
	}

	return &commitResolver{
		repo: repo,
		oid:  resolverstubs.GitObjectID(commitID),
		rev:  rev,
	}
}

func (r *commitResolver) ID() graphql.ID {
	return resolverstubs.MarshalID("GitCommit", map[string]any{
		"r": r.repo.ID(),
		"c": r.oid,
	})
}

func (r *commitResolver) Repository() resolverstubs.RepositoryResolver { return r.repo }
func (r *commitResolver) OID() resolverstubs.GitObjectID               { return r.oid }
func (r *commitResolver) AbbreviatedOID() string                       { return string(r.oid)[:7] }
func (r *commitResolver) URL() string                                  { return fmt.Sprintf("/%s/-/commit/%s", r.repo.Name(), r.rev) }
func (r *commitResolver) URI() string                                  { return fmt.Sprintf("/%s@%s", r.repo.Name(), r.rev) }

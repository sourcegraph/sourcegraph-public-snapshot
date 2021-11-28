package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (r *catalogComponentResolver) Commits(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.GitCommitConnectionResolver, error) {
	repoResolver, err := r.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): how to ensure both follow *and* sorting of results merged from `git log` over
	// multiple paths? Which sort order (topo or date) and how is that handled when the results are
	// merged? Follow doesn't work for multiple paths (see `git log --help`, "--follow ... works
	// only for a single file"), so we can't do this all in 1 Git command.

	return gql.NewGitCommitConnectionResolver(r.db, repoResolver, gql.GitCommitConnectionArgs{
		RevisionRange: r.sourceCommit,
		Path:          &r.sourcePaths,
		First:         args.First,
	}), nil
}

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

	return gql.NewGitCommitConnectionResolver(r.db, repoResolver, gql.GitCommitConnectionArgs{
		RevisionRange: r.sourceCommit,
		Path:          &r.sourcePath,
		First:         args.First,
	}), nil
}

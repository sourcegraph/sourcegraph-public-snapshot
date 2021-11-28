package resolvers

import (
	"context"
	"sort"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *catalogComponentResolver) Commits(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.GitCommitConnectionResolver, error) {
	// TODO(sqs): how to ensure both follow *and* sorting of results merged from `git log` over
	// multiple paths? Which sort order (topo or date) and how is that handled when the results are
	// merged? Follow doesn't work for multiple paths (see `git log --help`, "--follow ... works
	// only for a single file"), so we can't do this all in 1 Git command.

	var combinedCommits []*gitdomain.Commit
	for _, sourcePath := range r.sourcePaths {
		isDir := true
		commits, err := git.Commits(ctx, api.RepoName(r.sourceRepo), git.CommitsOptions{
			Range:  r.sourceCommit,
			Path:   sourcePath,
			Follow: !isDir,
			N:      uint(args.GetFirst()),
		})
		if err != nil {
			return nil, err
		}
		combinedCommits = append(combinedCommits, commits...)
	}

	sort.Slice(combinedCommits, func(i, j int) bool {
		return combinedCommits[i].Author.Date.After(combinedCommits[j].Author.Date)
	})

	var hasNextPage bool
	if len(combinedCommits) > int(args.GetFirst()) {
		combinedCommits = combinedCommits[:int(args.GetFirst())]
		hasNextPage = true
	}

	repoResolver, err := r.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}
	commitResolvers := make([]*gql.GitCommitResolver, len(combinedCommits))
	for i, c := range combinedCommits {
		commitResolvers[i] = gql.NewGitCommitResolver(r.db, repoResolver, c.ID, c)
	}
	return gql.NewStaticGitCommitConnection(commitResolvers, nil, hasNextPage), nil
}

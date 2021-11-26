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

func (r *componentResolver) Commits(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.GitCommitConnectionResolver, error) {
	// TODO(sqs): how to ensure both follow *and* sorting of results merged from `git log` over
	// multiple paths? Which sort order (topo or date) and how is that handled when the results are
	// merged? Follow doesn't work for multiple paths (see `git log --help`, "--follow ... works
	// only for a single file"), so we can't do this all in 1 Git command.

	slocs, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	type commitInfo struct {
		*gitdomain.Commit
		repoResolver *gql.RepositoryResolver
	}
	var combinedCommits []commitInfo
	for _, sloc := range slocs {
		for _, path := range sloc.paths {
			isDir := true
			commits, err := git.Commits(ctx, sloc.repoName, git.CommitsOptions{
				Range:  string(sloc.commitID),
				Path:   path,
				Follow: !isDir,
				N:      uint(args.GetFirst()),
			})
			if err != nil {
				return nil, err
			}
			for _, commit := range commits {
				combinedCommits = append(combinedCommits, commitInfo{
					Commit:       commit,
					repoResolver: sloc.repo,
				})
			}
		}
	}

	sort.Slice(combinedCommits, func(i, j int) bool {
		return combinedCommits[i].Author.Date.After(combinedCommits[j].Author.Date)
	})

	// Remove duplicate commits (that touched multiple paths).
	keep := combinedCommits[:0]
	var (
		lastCommitID api.CommitID
		lastRepo     api.RepoName
	)
	for _, c := range combinedCommits {
		if c.ID == lastCommitID && c.repoResolver.RepoName() == lastRepo {
			continue
		}
		keep = append(keep, c)
		lastCommitID = c.ID
		lastRepo = c.repoResolver.RepoName()
	}
	combinedCommits = keep

	var hasNextPage bool
	if len(combinedCommits) > int(args.GetFirst()) {
		combinedCommits = combinedCommits[:int(args.GetFirst())]
		hasNextPage = true
	}

	commitResolvers := make([]*gql.GitCommitResolver, len(combinedCommits))
	for i, c := range combinedCommits {
		commitResolvers[i] = gql.NewGitCommitResolver(r.db, c.repoResolver, c.ID, c.Commit)
	}
	return gql.NewStaticGitCommitConnection(commitResolvers, nil, hasNextPage), nil
}

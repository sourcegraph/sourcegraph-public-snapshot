package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type GitCommitConnectionResolver interface {
	Nodes(context.Context) ([]*GitCommitResolver, error)
	TotalCount(context.Context) (*int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type GitCommitConnectionArgs struct {
	RevisionRange string
	First         *int32
	Query         *string
	Path          *string
	Follow        bool
	Author        *string
	After         *string
}

func NewGitCommitConnectionResolver(db database.DB, repo *RepositoryResolver, args GitCommitConnectionArgs) GitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		db:   db,
		repo: repo,
		args: args,
	}
}

type gitCommitConnectionResolver struct {
	db   database.DB
	repo *RepositoryResolver
	args GitCommitConnectionArgs

	// cache results because it is used by multiple fields
	once    sync.Once
	commits []*gitdomain.Commit
	err     error
}

func (r *gitCommitConnectionResolver) compute(ctx context.Context) ([]*gitdomain.Commit, error) {
	do := func() ([]*gitdomain.Commit, error) {
		var n int32
		if r.args.First != nil {
			n = *r.args.First
			n++ // fetch +1 additional result so we can determine if a next page exists
		}
		var query string
		if r.args.Query != nil {
			query = *r.args.Query
		}
		var path string
		if r.args.Path != nil {
			path = *r.args.Path
		}
		var author string
		if r.args.Author != nil {
			author = *r.args.Author
		}
		var after string
		if r.args.After != nil {
			after = *r.args.After
		}
		return git.Commits(ctx, r.repo.RepoName(), git.CommitsOptions{
			Range:        r.args.RevisionRange,
			N:            uint(n),
			MessageQuery: query,
			Author:       author,
			After:        after,
			Path:         path,
			Follow:       r.args.Follow,
		}, authz.DefaultSubRepoPermsChecker)
	}

	r.once.Do(func() { r.commits, r.err = do() })
	return r.commits, r.err
}

func (r *gitCommitConnectionResolver) Nodes(ctx context.Context) ([]*GitCommitResolver, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.args.First != nil && len(commits) > int(*r.args.First) {
		// Don't return +1 results, which is used to determine if next page exists.
		commits = commits[:*r.args.First]
	}

	resolvers := make([]*GitCommitResolver, len(commits))
	for i, commit := range commits {
		resolvers[i] = NewGitCommitResolver(r.db, r.repo, commit.ID, commit)
	}

	return resolvers, nil
}

func (r *gitCommitConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if r.args.First != nil {
		// Return indeterminate total count if the caller requested an incomplete list of commits
		// (which means we'd need an extra and expensive Git operation to determine the total
		// count). This is to avoid `totalCount` taking significantly longer than `nodes` to
		// compute, which would be unexpected to many API clients.
		return nil, nil
	}
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	n := int32(len(commits))
	return &n, nil
}

func (r *gitCommitConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// If we have a limit, so we rely on having fetched +1 additional result in our limit to
	// indicate whether or not a next page exists.
	return graphqlutil.HasNextPage(r.args.First != nil && len(commits) > 0 && len(commits) > int(*r.args.First)), nil
}

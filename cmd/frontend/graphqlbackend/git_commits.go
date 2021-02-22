package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type gitCommitConnectionResolver struct {
	db            dbutil.DB
	revisionRange string

	first  *int32
	query  *string
	path   *string
	author *string
	after  *string

	repo *RepositoryResolver

	// cache results because it is used by multiple fields
	once    sync.Once
	commits []*git.Commit
	err     error
}

func (r *gitCommitConnectionResolver) compute(ctx context.Context) ([]*git.Commit, error) {
	do := func() ([]*git.Commit, error) {
		var n int32
		if r.first != nil {
			n = *r.first
			n++ // fetch +1 additional result so we can determine if a next page exists
		}
		var query string
		if r.query != nil {
			query = *r.query
		}
		var path string
		if r.path != nil {
			path = *r.path
		}
		var author string
		if r.author != nil {
			author = *r.author
		}
		var after string
		if r.after != nil {
			after = *r.after
		}
		return git.Commits(ctx, r.repo.name, git.CommitsOptions{
			Range:        r.revisionRange,
			N:            uint(n),
			MessageQuery: query,
			Author:       author,
			After:        after,
			Path:         path,
		})
	}

	r.once.Do(func() { r.commits, r.err = do() })
	return r.commits, r.err
}

func (r *gitCommitConnectionResolver) Nodes(ctx context.Context) ([]*GitCommitResolver, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(commits) > int(*r.first) {
		// Don't return +1 results, which is used to determine if next page exists.
		commits = commits[:*r.first]
	}

	resolvers := make([]*GitCommitResolver, len(commits))
	for i, commit := range commits {
		resolvers[i] = toGitCommitResolver(r.repo, r.db, commit.ID, commit)
	}

	return resolvers, nil
}

func (r *gitCommitConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if r.first != nil {
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
	return graphqlutil.HasNextPage(r.first != nil && len(commits) > 0 && len(commits) > int(*r.first)), nil
}

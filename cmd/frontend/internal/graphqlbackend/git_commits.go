package graphqlbackend

import (
	"context"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type gitCommitConnectionResolver struct {
	range_ string

	first *int32
	query *string
	path  *string

	repo *repositoryResolver

	// cache results because it is used by multiple fields
	once    sync.Once
	commits []*vcs.Commit
	err     error
}

func (r *gitCommitConnectionResolver) compute(ctx context.Context) ([]*vcs.Commit, error) {
	do := func() ([]*vcs.Commit, error) {
		vcsrepo := backend.Repos.CachedVCS(r.repo.repo)

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
		return vcsrepo.Commits(ctx, vcs.CommitsOptions{
			Range:        r.range_,
			N:            uint(n),
			MessageQuery: query,
			Path:         path,
		})
	}

	r.once.Do(func() { r.commits, r.err = do() })
	return r.commits, r.err
}

func (r *gitCommitConnectionResolver) Nodes(ctx context.Context) ([]*gitCommitResolver, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(commits) > int(*r.first) {
		// Don't return +1 results, which is used to determine if next page exists.
		commits = commits[:*r.first]
	}

	resolvers := make([]*gitCommitResolver, len(commits))
	for i, commit := range commits {
		resolvers[i] = toGitCommitResolver(r.repo, commit)
	}

	return resolvers, nil
}

func (r *gitCommitConnectionResolver) PageInfo(ctx context.Context) (*pageInfo, error) {
	commits, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// If we have a limit, so we rely on having fetched +1 additional result in our limit to
	// indicate whether or not a next page exists.
	return &pageInfo{
		hasNextPage: r.first != nil && len(commits) > 0 && len(commits) > int(*r.first),
	}, nil
}

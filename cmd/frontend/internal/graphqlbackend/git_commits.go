package graphqlbackend

import (
	"context"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type gitCommitConnectionResolver struct {
	headCommitID string

	first *int32
	query *string

	repo *repositoryResolver

	// cache results because it is used by multiple fields
	once    sync.Once
	commits []*vcs.Commit
	err     error
}

func (r *gitCommitConnectionResolver) compute(ctx context.Context) ([]*vcs.Commit, error) {
	do := func() ([]*vcs.Commit, error) {
		vcsrepo, err := db.RepoVCS.Open(ctx, r.repo.repo.ID)
		if err != nil {
			return nil, err
		}

		var n int32
		if r.first != nil {
			n = *r.first
		}
		var query string
		if r.query != nil {
			query = *r.query
		}
		return vcsrepo.Commits(ctx, vcs.CommitsOptions{
			Head:         vcs.CommitID(r.headCommitID),
			N:            uint(n),
			MessageQuery: query,
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

	// If the last commit in the list has parents, then there is another page.
	return &pageInfo{
		hasNextPage: len(commits) > 0 && len(commits[len(commits)-1].Parents) > 0,
	}, nil
}

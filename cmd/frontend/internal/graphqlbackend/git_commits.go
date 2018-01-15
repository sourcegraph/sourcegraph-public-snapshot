package graphqlbackend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type gitCommitConnectionResolver struct {
	headCommitID string

	first *int32
	query *string

	repo *repositoryResolver
}

func (r *gitCommitConnectionResolver) Nodes(ctx context.Context) ([]*gitCommitResolver, error) {
	vcsrepo, err := db.RepoVCS.Open(ctx, r.repo.repo.ID)
	if err != nil {
		return nil, err
	}

	var n int32
	if r.first != nil {
		n = *r.first
	}
	commits, _, err := vcsrepo.Commits(ctx, vcs.CommitsOptions{
		Head:    vcs.CommitID(r.headCommitID),
		N:       uint(n),
		NoTotal: true,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*gitCommitResolver, len(commits))
	for i, commit := range commits {
		resolvers[i] = toGitCommitResolver(r.repo, commit)
	}

	return resolvers, nil
}

func (r *gitCommitConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// TODO(sqs)
	return 0, errors.New("GitCommitConnection.totalCount is not yet supported")
}

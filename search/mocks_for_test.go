package search

import (
	"errors"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func reposGetNone(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
	return nil, errors.New("x")
}

func reposGetCommitOK(ctx context.Context, repoRevSpec *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
	return &vcs.Commit{ID: "c"}, nil
}

func reposGetBuildOK(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
	b := &sourcegraph.Build{CommitID: "c"}
	return &sourcegraph.RepoBuildInfo{
		Exact:          b,
		LastSuccessful: b,
	}, nil
}

func reposGetBuildOld(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
	return &sourcegraph.RepoBuildInfo{
		LastSuccessfulCommit: &vcs.Commit{
			ID: "c2"},

		CommitsBehind: 1,
	}, nil
}

func reposGetBuildNone(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
	return nil, errors.New("x")
}
